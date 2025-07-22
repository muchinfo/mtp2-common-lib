package socket

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"sync"
	"time"
)

// TCPServer TCP服务器结构体
type TCPServer struct {
	address            string                          // 监听地址
	listener           net.Listener                    // TCP监听器
	running            bool                            // 运行状态
	clients            map[string]*ClientConnection    // 客户端连接映射
	clientsMutex       sync.RWMutex                    // 客户端连接锁
	readTimeout        time.Duration                   // 读取超时
	writeTimeout       time.Duration                   // 写入超时
	maxConnections     int                             // 最大连接数
	ctx                context.Context                 // 上下文
	cancel             context.CancelFunc              // 取消函数
	wg                 sync.WaitGroup                  // 等待组
	onClientConnect    func(*ClientConnection)         // 客户端连接回调
	onClientDisconnect func(*ClientConnection, error)  // 客户端断开回调
	onMessage          func(*ClientConnection, []byte) // 消息接收回调
	onError            func(error)                     // 错误回调
}

// ClientConnection 客户端连接结构体
type ClientConnection struct {
	ID          string       // 连接ID
	Conn        net.Conn     // TCP连接
	RemoteAddr  string       // 远程地址
	ConnectedAt time.Time    // 连接时间
	server      *TCPServer   // 服务器引用
	mutex       sync.RWMutex // 读写锁
	closed      bool         // 是否已关闭
}

// TCPServerConfig TCP服务器配置
type TCPServerConfig struct {
	Address        string        // 监听地址，格式：:port 或 host:port
	ReadTimeout    time.Duration // 读取超时，默认30秒
	WriteTimeout   time.Duration // 写入超时，默认10秒
	MaxConnections int           // 最大连接数，0表示无限制
}

// NewTCPServer 创建新的TCP服务器
func NewTCPServer(config TCPServerConfig) *TCPServer {
	if config.ReadTimeout == 0 {
		config.ReadTimeout = 30 * time.Second
	}
	if config.WriteTimeout == 0 {
		config.WriteTimeout = 10 * time.Second
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &TCPServer{
		address:        config.Address,
		clients:        make(map[string]*ClientConnection),
		readTimeout:    config.ReadTimeout,
		writeTimeout:   config.WriteTimeout,
		maxConnections: config.MaxConnections,
		ctx:            ctx,
		cancel:         cancel,
	}
}

// SetCallbacks 设置回调函数
func (s *TCPServer) SetCallbacks(
	onClientConnect func(*ClientConnection),
	onClientDisconnect func(*ClientConnection, error),
	onMessage func(*ClientConnection, []byte),
	onError func(error),
) {
	s.clientsMutex.Lock()
	defer s.clientsMutex.Unlock()

	s.onClientConnect = onClientConnect
	s.onClientDisconnect = onClientDisconnect
	s.onMessage = onMessage
	s.onError = onError
}

// Start 启动服务器
func (s *TCPServer) Start() error {
	s.clientsMutex.Lock()
	if s.running {
		s.clientsMutex.Unlock()
		return fmt.Errorf("server is already running")
	}
	s.running = true
	s.clientsMutex.Unlock()

	listener, err := net.Listen("tcp", s.address)
	if err != nil {
		s.clientsMutex.Lock()
		s.running = false
		s.clientsMutex.Unlock()
		return fmt.Errorf("failed to start server on %s: %w", s.address, err)
	}

	s.listener = listener

	// 启动接受连接的goroutine
	s.wg.Add(1)
	go s.acceptLoop()

	return nil
}

// Stop 停止服务器
func (s *TCPServer) Stop() error {
	s.clientsMutex.Lock()
	if !s.running {
		s.clientsMutex.Unlock()
		return nil
	}
	s.running = false
	s.clientsMutex.Unlock()

	// 取消上下文
	s.cancel()

	// 关闭监听器
	if s.listener != nil {
		s.listener.Close()
	}

	// 关闭所有客户端连接
	s.clientsMutex.Lock()
	for _, client := range s.clients {
		client.Close()
	}
	s.clientsMutex.Unlock()

	// 等待所有goroutine结束
	s.wg.Wait()

	return nil
}

// IsRunning 检查服务器是否正在运行
func (s *TCPServer) IsRunning() bool {
	s.clientsMutex.RLock()
	defer s.clientsMutex.RUnlock()
	return s.running
}

// GetAddress 获取监听地址
func (s *TCPServer) GetAddress() string {
	if s.listener != nil {
		return s.listener.Addr().String()
	}
	return s.address
}

// GetClientCount 获取当前客户端连接数
func (s *TCPServer) GetClientCount() int {
	s.clientsMutex.RLock()
	defer s.clientsMutex.RUnlock()
	return len(s.clients)
}

// GetClients 获取所有客户端连接
func (s *TCPServer) GetClients() []*ClientConnection {
	s.clientsMutex.RLock()
	defer s.clientsMutex.RUnlock()

	clients := make([]*ClientConnection, 0, len(s.clients))
	for _, client := range s.clients {
		clients = append(clients, client)
	}
	return clients
}

// GetClient 根据ID获取客户端连接
func (s *TCPServer) GetClient(id string) *ClientConnection {
	s.clientsMutex.RLock()
	defer s.clientsMutex.RUnlock()
	return s.clients[id]
}

// Broadcast 向所有客户端广播消息
func (s *TCPServer) Broadcast(data []byte) {
	s.clientsMutex.RLock()
	clients := make([]*ClientConnection, 0, len(s.clients))
	for _, client := range s.clients {
		clients = append(clients, client)
	}
	s.clientsMutex.RUnlock()

	for _, client := range clients {
		if err := client.Send(data); err != nil && s.onError != nil {
			go s.onError(fmt.Errorf("failed to broadcast to client %s: %w", client.ID, err))
		}
	}
}

// BroadcastString 向所有客户端广播字符串消息
func (s *TCPServer) BroadcastString(message string) {
	s.Broadcast([]byte(message))
}

// acceptLoop 接受连接循环
func (s *TCPServer) acceptLoop() {
	defer s.wg.Done()

	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			conn, err := s.listener.Accept()
			if err != nil {
				select {
				case <-s.ctx.Done():
					return
				default:
					if s.onError != nil {
						go s.onError(fmt.Errorf("failed to accept connection: %w", err))
					}
					continue
				}
			}

			// 检查连接数限制
			s.clientsMutex.RLock()
			currentConnections := len(s.clients)
			s.clientsMutex.RUnlock()

			if s.maxConnections > 0 && currentConnections >= s.maxConnections {
				conn.Close()
				if s.onError != nil {
					go s.onError(fmt.Errorf("connection limit reached (%d), rejecting connection from %s", s.maxConnections, conn.RemoteAddr()))
				}
				continue
			}

			// 创建客户端连接
			client := s.newClientConnection(conn)

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
	}
}

// newClientConnection 创建新的客户端连接
func (s *TCPServer) newClientConnection(conn net.Conn) *ClientConnection {
	return &ClientConnection{
		ID:          generateConnectionID(conn),
		Conn:        conn,
		RemoteAddr:  conn.RemoteAddr().String(),
		ConnectedAt: time.Now(),
		server:      s,
	}
}

// handleClient 处理客户端连接
func (s *TCPServer) handleClient(client *ClientConnection) {
	defer func() {
		s.wg.Done()
		s.removeClient(client)
	}()

	reader := bufio.NewReader(client.Conn)
	buffer := make([]byte, 4096)

	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			// 设置读取超时
			if s.readTimeout > 0 {
				client.Conn.SetReadDeadline(time.Now().Add(s.readTimeout))
			}

			n, err := reader.Read(buffer)
			if err != nil {
				client.Close()
				if s.onClientDisconnect != nil {
					go s.onClientDisconnect(client, err)
				}
				return
			}

			if n > 0 && s.onMessage != nil {
				// 复制数据避免竞态条件
				data := make([]byte, n)
				copy(data, buffer[:n])
				go s.onMessage(client, data)
			}
		}
	}
}

// removeClient 移除客户端连接
func (s *TCPServer) removeClient(client *ClientConnection) {
	s.clientsMutex.Lock()
	delete(s.clients, client.ID)
	s.clientsMutex.Unlock()
}

// Send 发送数据到客户端
func (c *ClientConnection) Send(data []byte) error {
	c.mutex.RLock()
	if c.closed {
		c.mutex.RUnlock()
		return fmt.Errorf("connection is closed")
	}
	conn := c.Conn
	c.mutex.RUnlock()

	// 设置写入超时
	if c.server.writeTimeout > 0 {
		conn.SetWriteDeadline(time.Now().Add(c.server.writeTimeout))
	}

	_, err := conn.Write(data)
	if err != nil {
		c.Close()
		return fmt.Errorf("failed to send data to client %s: %w", c.ID, err)
	}

	return nil
}

// SendString 发送字符串数据到客户端
func (c *ClientConnection) SendString(message string) error {
	return c.Send([]byte(message))
}

// Close 关闭客户端连接
func (c *ClientConnection) Close() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.closed {
		c.closed = true
		if c.Conn != nil {
			c.Conn.Close()
		}
	}
}

// IsClosed 检查连接是否已关闭
func (c *ClientConnection) IsClosed() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.closed
}

// GetUptime 获取连接持续时间
func (c *ClientConnection) GetUptime() time.Duration {
	return time.Since(c.ConnectedAt)
}

// generateConnectionID 生成连接ID
func generateConnectionID(conn net.Conn) string {
	return fmt.Sprintf("%s_%d", conn.RemoteAddr().String(), time.Now().UnixNano())
}
