package websocket

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WSServer WebSocket服务器结构体
type WSServer struct {
	address            string                            // 配置的监听地址
	actualAddress      string                            // 实际监听地址
	path               string                            // WebSocket路径
	server             *http.Server                      // HTTP服务器
	upgrader           websocket.Upgrader                // WebSocket升级器
	clients            map[string]*WSClientConnection    // 客户端连接映射
	clientsMutex       sync.RWMutex                      // 客户端连接锁
	running            bool                              // 运行状态
	maxConnections     int                               // 最大连接数
	pingInterval       time.Duration                     // ping间隔
	pongWait           time.Duration                     // pong等待时间
	writeWait          time.Duration                     // 写入等待时间
	readBufferSize     int                               // 读取缓冲区大小
	writeBufferSize    int                               // 写入缓冲区大小
	ctx                context.Context                   // 上下文
	cancel             context.CancelFunc                // 取消函数
	wg                 sync.WaitGroup                    // 等待组
	onClientConnect    func(*WSClientConnection)         // 客户端连接回调
	onClientDisconnect func(*WSClientConnection, error)  // 客户端断开回调
	onMessage          func(*WSClientConnection, []byte) // 消息接收回调
	onError            func(error)                       // 错误回调
}

// WSClientConnection WebSocket客户端连接结构体
type WSClientConnection struct {
	ID          string             // 连接ID
	Conn        *websocket.Conn    // WebSocket连接
	RemoteAddr  string             // 远程地址
	ConnectedAt time.Time          // 连接时间
	UserAgent   string             // 用户代理
	Headers     http.Header        // HTTP头
	server      *WSServer          // 服务器引用
	mutex       sync.RWMutex       // 读写锁
	closed      bool               // 是否已关闭
	ctx         context.Context    // 上下文
	cancel      context.CancelFunc // 取消函数
}

// WSServerConfig WebSocket服务器配置
type WSServerConfig struct {
	Address         string                     // 监听地址，格式：:port 或 host:port
	Path            string                     // WebSocket路径，默认"/"
	MaxConnections  int                        // 最大连接数，0表示无限制
	PingInterval    time.Duration              // ping间隔，默认30秒
	PongWait        time.Duration              // pong等待时间，默认60秒
	WriteWait       time.Duration              // 写入等待时间，默认10秒
	ReadBufferSize  int                        // 读取缓冲区大小，默认4096
	WriteBufferSize int                        // 写入缓冲区大小，默认4096
	CheckOrigin     func(r *http.Request) bool // 跨域检查函数
}

// NewWSServer 创建新的WebSocket服务器
func NewWSServer(config WSServerConfig) *WSServer {
	// 设置默认值
	if config.Path == "" {
		config.Path = "/"
	}
	if config.PingInterval == 0 {
		config.PingInterval = 30 * time.Second
	}
	if config.PongWait == 0 {
		config.PongWait = 60 * time.Second
	}
	if config.WriteWait == 0 {
		config.WriteWait = 10 * time.Second
	}
	if config.ReadBufferSize == 0 {
		config.ReadBufferSize = 4096
	}
	if config.WriteBufferSize == 0 {
		config.WriteBufferSize = 4096
	}

	ctx, cancel := context.WithCancel(context.Background())

	// 配置WebSocket升级器
	upgrader := websocket.Upgrader{
		ReadBufferSize:  config.ReadBufferSize,
		WriteBufferSize: config.WriteBufferSize,
		CheckOrigin:     config.CheckOrigin,
	}

	// 如果没有提供CheckOrigin函数，使用默认的（允许所有来源）
	if upgrader.CheckOrigin == nil {
		upgrader.CheckOrigin = func(r *http.Request) bool {
			return true
		}
	}

	return &WSServer{
		address:         config.Address,
		path:            config.Path,
		upgrader:        upgrader,
		clients:         make(map[string]*WSClientConnection),
		maxConnections:  config.MaxConnections,
		pingInterval:    config.PingInterval,
		pongWait:        config.PongWait,
		writeWait:       config.WriteWait,
		readBufferSize:  config.ReadBufferSize,
		writeBufferSize: config.WriteBufferSize,
		ctx:             ctx,
		cancel:          cancel,
	}
}

// SetCallbacks 设置回调函数
func (s *WSServer) SetCallbacks(
	onClientConnect func(*WSClientConnection),
	onClientDisconnect func(*WSClientConnection, error),
	onMessage func(*WSClientConnection, []byte),
	onError func(error),
) {
	s.clientsMutex.Lock()
	defer s.clientsMutex.Unlock()

	s.onClientConnect = onClientConnect
	s.onClientDisconnect = onClientDisconnect
	s.onMessage = onMessage
	s.onError = onError
}

// Start 启动WebSocket服务器
func (s *WSServer) Start() error {
	s.clientsMutex.Lock()
	if s.running {
		s.clientsMutex.Unlock()
		return fmt.Errorf("server is already running")
	}
	s.running = true
	s.clientsMutex.Unlock()

	// 创建监听器
	listener, err := net.Listen("tcp", s.address)
	if err != nil {
		s.clientsMutex.Lock()
		s.running = false
		s.clientsMutex.Unlock()
		return fmt.Errorf("failed to listen on %s: %w", s.address, err)
	}

	// 保存实际的监听地址
	s.actualAddress = listener.Addr().String()

	// 创建HTTP服务器
	mux := http.NewServeMux()
	mux.HandleFunc(s.path, s.handleWebSocket)

	s.server = &http.Server{
		Handler: mux,
	}

	// 启动服务器
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		if err := s.server.Serve(listener); err != nil && err != http.ErrServerClosed {
			if s.onError != nil {
				s.onError(fmt.Errorf("server listen error: %w", err))
			}
		}
	}()

	return nil
}

// Stop 停止WebSocket服务器
func (s *WSServer) Stop() error {
	s.clientsMutex.Lock()
	if !s.running {
		s.clientsMutex.Unlock()
		return nil
	}
	s.running = false
	s.clientsMutex.Unlock()

	// 取消上下文
	s.cancel()

	// 关闭所有客户端连接
	s.clientsMutex.Lock()
	for _, client := range s.clients {
		client.Close()
	}
	s.clientsMutex.Unlock()

	// 关闭HTTP服务器
	if s.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		s.server.Shutdown(ctx)
	}

	// 等待所有goroutine结束
	s.wg.Wait()

	return nil
}

// IsRunning 检查服务器是否正在运行
func (s *WSServer) IsRunning() bool {
	s.clientsMutex.RLock()
	defer s.clientsMutex.RUnlock()
	return s.running
}

// GetAddress 获取实际监听地址
func (s *WSServer) GetAddress() string {
	if s.actualAddress != "" {
		return s.actualAddress
	}
	return s.address
}

// GetClientCount 获取当前客户端连接数
func (s *WSServer) GetClientCount() int {
	s.clientsMutex.RLock()
	defer s.clientsMutex.RUnlock()
	return len(s.clients)
}

// GetClients 获取所有客户端连接
func (s *WSServer) GetClients() []*WSClientConnection {
	s.clientsMutex.RLock()
	defer s.clientsMutex.RUnlock()

	clients := make([]*WSClientConnection, 0, len(s.clients))
	for _, client := range s.clients {
		clients = append(clients, client)
	}
	return clients
}

// GetClient 根据ID获取客户端连接
func (s *WSServer) GetClient(id string) *WSClientConnection {
	s.clientsMutex.RLock()
	defer s.clientsMutex.RUnlock()
	return s.clients[id]
}

// Broadcast 向所有客户端广播二进制数据
func (s *WSServer) Broadcast(data []byte) {
	s.BroadcastMessage(websocket.BinaryMessage, data)
}

// BroadcastText 向所有客户端广播文本消息
func (s *WSServer) BroadcastText(text string) {
	s.BroadcastMessage(websocket.TextMessage, []byte(text))
}

// BroadcastJSON 向所有客户端广播JSON消息
func (s *WSServer) BroadcastJSON(v interface{}) {
	s.clientsMutex.RLock()
	clients := make([]*WSClientConnection, 0, len(s.clients))
	for _, client := range s.clients {
		clients = append(clients, client)
	}
	s.clientsMutex.RUnlock()

	for _, client := range clients {
		if err := client.SendJSON(v); err != nil && s.onError != nil {
			go s.onError(fmt.Errorf("failed to broadcast JSON to client %s: %w", client.ID, err))
		}
	}
}

// BroadcastMessage 向所有客户端广播指定类型的消息
func (s *WSServer) BroadcastMessage(messageType int, data []byte) {
	s.clientsMutex.RLock()
	clients := make([]*WSClientConnection, 0, len(s.clients))
	for _, client := range s.clients {
		clients = append(clients, client)
	}
	s.clientsMutex.RUnlock()

	for _, client := range clients {
		if err := client.SendMessage(messageType, data); err != nil && s.onError != nil {
			go s.onError(fmt.Errorf("failed to broadcast to client %s: %w", client.ID, err))
		}
	}
}

// handleWebSocket 处理WebSocket连接升级
func (s *WSServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// 检查连接数限制
	s.clientsMutex.RLock()
	currentConnections := len(s.clients)
	s.clientsMutex.RUnlock()

	if s.maxConnections > 0 && currentConnections >= s.maxConnections {
		http.Error(w, "Connection limit reached", http.StatusServiceUnavailable)
		return
	}

	// 升级为WebSocket连接
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		if s.onError != nil {
			go s.onError(fmt.Errorf("websocket upgrade error: %w", err))
		}
		return
	}

	// 创建客户端连接
	client := s.newClientConnection(conn, r)

	// 添加到客户端映射
	s.clientsMutex.Lock()
	s.clients[client.ID] = client
	s.clientsMutex.Unlock()

	// 启动客户端处理goroutine
	s.wg.Add(1)
	go s.handleClient(client)

	// 触发客户端连接回调
	if s.onClientConnect != nil {
		go s.onClientConnect(client)
	}
}

// newClientConnection 创建新的客户端连接
func (s *WSServer) newClientConnection(conn *websocket.Conn, r *http.Request) *WSClientConnection {
	ctx, cancel := context.WithCancel(s.ctx)

	return &WSClientConnection{
		ID:          generateWSConnectionID(conn, r),
		Conn:        conn,
		RemoteAddr:  conn.RemoteAddr().String(),
		ConnectedAt: time.Now(),
		UserAgent:   r.UserAgent(),
		Headers:     r.Header.Clone(),
		server:      s,
		ctx:         ctx,
		cancel:      cancel,
	}
}

// handleClient 处理客户端连接
func (s *WSServer) handleClient(client *WSClientConnection) {
	defer func() {
		s.wg.Done()
		s.removeClient(client)
		client.Close()
	}()

	// 设置连接参数
	client.Conn.SetReadDeadline(time.Now().Add(s.pongWait))
	client.Conn.SetPongHandler(func(string) error {
		client.Conn.SetReadDeadline(time.Now().Add(s.pongWait))
		return nil
	})

	// 启动ping goroutine
	go s.pingClient(client)

	// 消息读取循环
	for {
		select {
		case <-client.ctx.Done():
			return
		default:
			_, message, err := client.Conn.ReadMessage()
			if err != nil {
				if s.onClientDisconnect != nil {
					go s.onClientDisconnect(client, err)
				}
				return
			}

			if len(message) > 0 && s.onMessage != nil {
				// 复制数据避免竞态条件
				data := make([]byte, len(message))
				copy(data, message)
				go s.onMessage(client, data)
			}
		}
	}
}

// pingClient 向客户端发送ping消息
func (s *WSServer) pingClient(client *WSClientConnection) {
	ticker := time.NewTicker(s.pingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-client.ctx.Done():
			return
		case <-ticker.C:
			client.mutex.RLock()
			conn := client.Conn
			closed := client.closed
			client.mutex.RUnlock()

			if closed || conn == nil {
				return
			}

			conn.SetWriteDeadline(time.Now().Add(s.writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				client.Close()
				return
			}
		}
	}
}

// removeClient 移除客户端连接
func (s *WSServer) removeClient(client *WSClientConnection) {
	s.clientsMutex.Lock()
	delete(s.clients, client.ID)
	s.clientsMutex.Unlock()
}

// Send 发送二进制数据到客户端
func (c *WSClientConnection) Send(data []byte) error {
	return c.SendMessage(websocket.BinaryMessage, data)
}

// SendText 发送文本消息到客户端
func (c *WSClientConnection) SendText(text string) error {
	return c.SendMessage(websocket.TextMessage, []byte(text))
}

// SendJSON 发送JSON消息到客户端
func (c *WSClientConnection) SendJSON(v interface{}) error {
	c.mutex.RLock()
	if c.closed {
		c.mutex.RUnlock()
		return fmt.Errorf("connection is closed")
	}
	conn := c.Conn
	c.mutex.RUnlock()

	conn.SetWriteDeadline(time.Now().Add(c.server.writeWait))
	err := conn.WriteJSON(v)
	if err != nil {
		c.Close()
		return fmt.Errorf("failed to send JSON to client %s: %w", c.ID, err)
	}

	return nil
}

// SendMessage 发送指定类型的消息到客户端
func (c *WSClientConnection) SendMessage(messageType int, data []byte) error {
	c.mutex.RLock()
	if c.closed {
		c.mutex.RUnlock()
		return fmt.Errorf("connection is closed")
	}
	conn := c.Conn
	c.mutex.RUnlock()

	conn.SetWriteDeadline(time.Now().Add(c.server.writeWait))
	err := conn.WriteMessage(messageType, data)
	if err != nil {
		c.Close()
		return fmt.Errorf("failed to send message to client %s: %w", c.ID, err)
	}

	return nil
}

// Close 关闭客户端连接
func (c *WSClientConnection) Close() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.closed {
		c.closed = true
		c.cancel()
		if c.Conn != nil {
			// 发送关闭消息
			c.Conn.WriteMessage(websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			c.Conn.Close()
		}
	}
}

// IsClosed 检查连接是否已关闭
func (c *WSClientConnection) IsClosed() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.closed
}

// GetUptime 获取连接持续时间
func (c *WSClientConnection) GetUptime() time.Duration {
	return time.Since(c.ConnectedAt)
}

// generateWSConnectionID 生成WebSocket连接ID
func generateWSConnectionID(conn *websocket.Conn, r *http.Request) string {
	return fmt.Sprintf("%s_%s_%d",
		conn.RemoteAddr().String(),
		r.UserAgent(),
		time.Now().UnixNano())
}
