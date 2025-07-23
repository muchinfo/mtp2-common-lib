package websocket

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WSClient WebSocket客户端结构体
type WSClient struct {
	url             string             // 服务器URL
	conn            *websocket.Conn    // WebSocket连接
	connected       bool               // 连接状态
	reconnectDelay  time.Duration      // 重连延迟
	maxReconnects   int                // 最大重连次数
	reconnectCount  int                // 当前重连次数
	autoReconnect   bool               // 是否自动重连
	headers         http.Header        // 连接时的HTTP头
	mutex           sync.RWMutex       // 读写锁
	ctx             context.Context    // 上下文
	cancel          context.CancelFunc // 取消函数
	onConnect       func()             // 连接成功回调
	onDisconnect    func(error)        // 断开连接回调
	onMessage       func([]byte)       // 消息接收回调
	onError         func(error)        // 错误回调
	pingInterval    time.Duration      // ping间隔
	pongWait        time.Duration      // pong等待时间
	writeWait       time.Duration      // 写入等待时间
	readBufferSize  int                // 读取缓冲区大小
	writeBufferSize int                // 写入缓冲区大小
}

// WSClientConfig WebSocket客户端配置
type WSClientConfig struct {
	URL             string        // 服务器URL，格式：ws://host:port/path 或 wss://host:port/path
	ReconnectDelay  time.Duration // 重连延迟，默认5秒
	MaxReconnects   int           // 最大重连次数，0表示无限重连
	AutoReconnect   bool          // 是否自动重连，默认true
	Headers         http.Header   // 连接时的HTTP头
	PingInterval    time.Duration // ping间隔，默认30秒
	PongWait        time.Duration // pong等待时间，默认60秒
	WriteWait       time.Duration // 写入等待时间，默认10秒
	ReadBufferSize  int           // 读取缓冲区大小，默认4096
	WriteBufferSize int           // 写入缓冲区大小，默认4096
}

// NewWSClient 创建新的WebSocket客户端
func NewWSClient(config WSClientConfig) *WSClient {
	// 设置默认值
	if config.ReconnectDelay == 0 {
		config.ReconnectDelay = 5 * time.Second
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
	if config.Headers == nil {
		config.Headers = make(http.Header)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &WSClient{
		url:             config.URL,
		reconnectDelay:  config.ReconnectDelay,
		maxReconnects:   config.MaxReconnects,
		autoReconnect:   config.AutoReconnect,
		headers:         config.Headers,
		ctx:             ctx,
		cancel:          cancel,
		pingInterval:    config.PingInterval,
		pongWait:        config.PongWait,
		writeWait:       config.WriteWait,
		readBufferSize:  config.ReadBufferSize,
		writeBufferSize: config.WriteBufferSize,
	}
}

// SetCallbacks 设置回调函数
func (c *WSClient) SetCallbacks(
	onConnect func(),
	onDisconnect func(error),
	onMessage func([]byte),
	onError func(error),
) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.onConnect = onConnect
	c.onDisconnect = onDisconnect
	c.onMessage = onMessage
	c.onError = onError
}

// Connect 连接到WebSocket服务器
func (c *WSClient) Connect() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.connected {
		return fmt.Errorf("already connected")
	}

	// 解析URL
	u, err := url.Parse(c.url)
	if err != nil {
		return fmt.Errorf("invalid URL %s: %w", c.url, err)
	}

	// 创建dialer
	dialer := websocket.Dialer{
		ReadBufferSize:   c.readBufferSize,
		WriteBufferSize:  c.writeBufferSize,
		HandshakeTimeout: 10 * time.Second,
	}

	// 建立WebSocket连接
	conn, _, err := dialer.Dial(u.String(), c.headers)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", c.url, err)
	}

	c.conn = conn
	c.connected = true
	c.reconnectCount = 0

	// 设置连接参数
	c.conn.SetReadDeadline(time.Now().Add(c.pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(c.pongWait))
		return nil
	})

	// 触发连接成功回调
	if c.onConnect != nil {
		go c.onConnect()
	}

	// 启动读取和ping goroutines
	go c.readLoop()
	go c.pingLoop()

	return nil
}

// Disconnect 断开连接
func (c *WSClient) Disconnect() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.connected {
		return
	}

	c.connected = false
	if c.conn != nil {
		// 发送关闭消息
		c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		c.conn.Close()
		c.conn = nil
	}

	// 触发断开连接回调
	if c.onDisconnect != nil {
		go c.onDisconnect(nil)
	}
}

// Close 关闭客户端
func (c *WSClient) Close() {
	c.cancel()
	c.Disconnect()
}

// Send 发送二进制数据
func (c *WSClient) Send(data []byte) error {
	return c.SendMessage(websocket.BinaryMessage, data)
}

// SendText 发送文本消息
func (c *WSClient) SendText(text string) error {
	return c.SendMessage(websocket.TextMessage, []byte(text))
}

// SendJSON 发送JSON消息
func (c *WSClient) SendJSON(v interface{}) error {
	c.mutex.RLock()
	conn := c.conn
	connected := c.connected
	c.mutex.RUnlock()

	if !connected || conn == nil {
		return fmt.Errorf("not connected")
	}

	conn.SetWriteDeadline(time.Now().Add(c.writeWait))
	return conn.WriteJSON(v)
}

// SendMessage 发送指定类型的消息
func (c *WSClient) SendMessage(messageType int, data []byte) error {
	c.mutex.RLock()
	conn := c.conn
	connected := c.connected
	c.mutex.RUnlock()

	if !connected || conn == nil {
		return fmt.Errorf("not connected")
	}

	conn.SetWriteDeadline(time.Now().Add(c.writeWait))
	err := conn.WriteMessage(messageType, data)
	if err != nil {
		c.handleConnectionError(err)
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

// IsConnected 检查连接状态
func (c *WSClient) IsConnected() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.connected
}

// GetURL 获取服务器URL
func (c *WSClient) GetURL() string {
	return c.url
}

// GetReconnectCount 获取重连次数
func (c *WSClient) GetReconnectCount() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.reconnectCount
}

// readLoop 读取消息循环
func (c *WSClient) readLoop() {
	defer func() {
		c.mutex.Lock()
		c.connected = false
		if c.conn != nil {
			c.conn.Close()
			c.conn = nil
		}
		c.mutex.Unlock()
	}()

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			_, message, err := c.conn.ReadMessage()
			if err != nil {
				c.handleConnectionError(err)
				return
			}

			if len(message) > 0 && c.onMessage != nil {
				// 复制数据避免竞态条件
				data := make([]byte, len(message))
				copy(data, message)
				go c.onMessage(data)
			}
		}
	}
}

// pingLoop ping循环
func (c *WSClient) pingLoop() {
	ticker := time.NewTicker(c.pingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			c.mutex.RLock()
			conn := c.conn
			connected := c.connected
			c.mutex.RUnlock()

			if !connected || conn == nil {
				return
			}

			conn.SetWriteDeadline(time.Now().Add(c.writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.handleConnectionError(err)
				return
			}
		}
	}
}

// handleConnectionError 处理连接错误
func (c *WSClient) handleConnectionError(err error) {
	c.mutex.Lock()
	wasConnected := c.connected
	c.connected = false
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
	c.mutex.Unlock()

	if wasConnected {
		// 触发断开连接回调
		if c.onDisconnect != nil {
			go c.onDisconnect(err)
		}

		// 触发错误回调
		if c.onError != nil {
			go c.onError(err)
		}

		// 自动重连
		if c.autoReconnect {
			go c.reconnect()
		}
	}
}

// reconnect 重连逻辑
func (c *WSClient) reconnect() {
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			// 检查是否达到最大重连次数
			c.mutex.RLock()
			currentCount := c.reconnectCount
			maxReconnects := c.maxReconnects
			c.mutex.RUnlock()

			if maxReconnects > 0 && currentCount >= maxReconnects {
				if c.onError != nil {
					go c.onError(fmt.Errorf("max reconnect attempts (%d) reached", maxReconnects))
				}
				return
			}

			// 等待重连延迟
			time.Sleep(c.reconnectDelay)

			// 尝试重连
			c.mutex.Lock()
			c.reconnectCount++
			c.mutex.Unlock()

			if err := c.Connect(); err != nil {
				if c.onError != nil {
					go c.onError(fmt.Errorf("reconnect attempt %d failed: %w", currentCount+1, err))
				}
				continue
			}

			return
		}
	}
}
