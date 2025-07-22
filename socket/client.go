package socket

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"sync"
	"time"
)

// TCPClient TCP客户端结构体
type TCPClient struct {
	address        string        // 服务器地址
	conn           net.Conn      // TCP连接
	connected      bool          // 连接状态
	reconnectDelay time.Duration // 重连延迟
	readTimeout    time.Duration // 读取超时
	writeTimeout   time.Duration // 写入超时
	maxReconnects  int           // 最大重连次数
	reconnectCount int           // 当前重连次数
	autoReconnect  bool          // 是否自动重连
	mutex          sync.RWMutex  // 读写锁
	ctx            context.Context
	cancel         context.CancelFunc
	onConnect      func()       // 连接成功回调
	onDisconnect   func(error)  // 断开连接回调
	onMessage      func([]byte) // 消息接收回调
	onError        func(error)  // 错误回调
}

// TCPClientConfig TCP客户端配置
type TCPClientConfig struct {
	Address        string        // 服务器地址，格式：host:port
	ReconnectDelay time.Duration // 重连延迟，默认5秒
	ReadTimeout    time.Duration // 读取超时，默认30秒
	WriteTimeout   time.Duration // 写入超时，默认10秒
	MaxReconnects  int           // 最大重连次数，0表示无限重连
	AutoReconnect  bool          // 是否自动重连，默认true
}

// NewTCPClient 创建新的TCP客户端
func NewTCPClient(config TCPClientConfig) *TCPClient {
	if config.ReconnectDelay == 0 {
		config.ReconnectDelay = 5 * time.Second
	}
	if config.ReadTimeout == 0 {
		config.ReadTimeout = 30 * time.Second
	}
	if config.WriteTimeout == 0 {
		config.WriteTimeout = 10 * time.Second
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &TCPClient{
		address:        config.Address,
		reconnectDelay: config.ReconnectDelay,
		readTimeout:    config.ReadTimeout,
		writeTimeout:   config.WriteTimeout,
		maxReconnects:  config.MaxReconnects,
		autoReconnect:  config.AutoReconnect,
		ctx:            ctx,
		cancel:         cancel,
	}
}

// SetCallbacks 设置回调函数
func (c *TCPClient) SetCallbacks(
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

// Connect 连接到服务器
func (c *TCPClient) Connect() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.connected {
		return fmt.Errorf("already connected")
	}

	conn, err := net.DialTimeout("tcp", c.address, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", c.address, err)
	}

	c.conn = conn
	c.connected = true
	c.reconnectCount = 0

	// 触发连接成功回调
	if c.onConnect != nil {
		go c.onConnect()
	}

	// 启动读取goroutine
	go c.readLoop()

	return nil
}

// Disconnect 断开连接
func (c *TCPClient) Disconnect() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.connected {
		return
	}

	c.connected = false
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}

	// 触发断开连接回调
	if c.onDisconnect != nil {
		go c.onDisconnect(nil)
	}
}

// Close 关闭客户端
func (c *TCPClient) Close() {
	c.cancel()
	c.Disconnect()
}

// Send 发送数据
func (c *TCPClient) Send(data []byte) error {
	c.mutex.RLock()
	conn := c.conn
	connected := c.connected
	c.mutex.RUnlock()

	if !connected || conn == nil {
		return fmt.Errorf("not connected")
	}

	// 设置写入超时
	if c.writeTimeout > 0 {
		conn.SetWriteDeadline(time.Now().Add(c.writeTimeout))
	}

	_, err := conn.Write(data)
	if err != nil {
		c.handleConnectionError(err)
		return fmt.Errorf("failed to send data: %w", err)
	}

	return nil
}

// SendString 发送字符串数据
func (c *TCPClient) SendString(message string) error {
	return c.Send([]byte(message))
}

// IsConnected 检查连接状态
func (c *TCPClient) IsConnected() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.connected
}

// GetAddress 获取服务器地址
func (c *TCPClient) GetAddress() string {
	return c.address
}

// GetReconnectCount 获取重连次数
func (c *TCPClient) GetReconnectCount() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.reconnectCount
}

// readLoop 读取循环
func (c *TCPClient) readLoop() {
	defer func() {
		c.mutex.Lock()
		c.connected = false
		if c.conn != nil {
			c.conn.Close()
			c.conn = nil
		}
		c.mutex.Unlock()
	}()

	reader := bufio.NewReader(c.conn)
	buffer := make([]byte, 4096)

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			// 设置读取超时
			if c.readTimeout > 0 {
				c.conn.SetReadDeadline(time.Now().Add(c.readTimeout))
			}

			n, err := reader.Read(buffer)
			if err != nil {
				c.handleConnectionError(err)
				return
			}

			if n > 0 && c.onMessage != nil {
				// 复制数据避免竞态条件
				data := make([]byte, n)
				copy(data, buffer[:n])
				go c.onMessage(data)
			}
		}
	}
}

// handleConnectionError 处理连接错误
func (c *TCPClient) handleConnectionError(err error) {
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
func (c *TCPClient) reconnect() {
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
