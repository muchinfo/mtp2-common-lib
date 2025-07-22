package socket

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestTCPServer_StartStop(t *testing.T) {
	config := TCPServerConfig{
		Address:        ":0", // 使用随机端口
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		MaxConnections: 10,
	}

	server := NewTCPServer(config)

	// 测试启动
	err := server.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	// 检查运行状态
	if !server.IsRunning() {
		t.Error("Server should be running")
	}

	// 获取实际监听地址
	address := server.GetAddress()
	if address == "" {
		t.Error("Server address should not be empty")
	}

	// 测试停止
	err = server.Stop()
	if err != nil {
		t.Fatalf("Failed to stop server: %v", err)
	}

	// 检查运行状态
	if server.IsRunning() {
		t.Error("Server should not be running")
	}
}

func TestTCPServer_ClientConnection(t *testing.T) {
	config := TCPServerConfig{
		Address:        ":0",
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		MaxConnections: 5,
	}

	server := NewTCPServer(config)
	defer server.Stop()

	var wg sync.WaitGroup
	var connectedClient *ClientConnection

	// 设置回调函数
	server.SetCallbacks(
		func(client *ClientConnection) {
			connectedClient = client
			wg.Done()
		},
		func(client *ClientConnection, err error) {
			fmt.Printf("Client disconnected: %s, error: %v\n", client.ID, err)
		},
		func(client *ClientConnection, data []byte) {
			fmt.Printf("Received from %s: %s\n", client.ID, string(data))
		},
		func(err error) {
			fmt.Printf("Server error: %v\n", err)
		},
	)

	// 启动服务器
	err := server.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	// 连接到服务器
	wg.Add(1)
	conn, err := net.Dial("tcp", server.GetAddress())
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	// 等待连接建立
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// 检查客户端连接
		if connectedClient == nil {
			t.Error("Connected client should not be nil")
		}
		if connectedClient.RemoteAddr == "" {
			t.Error("Client remote address should not be empty")
		}
		if connectedClient.IsClosed() {
			t.Error("Client should not be closed")
		}

		// 检查服务器客户端计数
		if server.GetClientCount() != 1 {
			t.Errorf("Expected 1 client, got %d", server.GetClientCount())
		}

		// 获取客户端列表
		clients := server.GetClients()
		if len(clients) != 1 {
			t.Errorf("Expected 1 client in list, got %d", len(clients))
		}

		// 根据ID获取客户端
		client := server.GetClient(connectedClient.ID)
		if client == nil {
			t.Error("Should be able to get client by ID")
		}

	case <-time.After(5 * time.Second):
		t.Error("Timeout waiting for client connection")
	}
}

func TestTCPServer_MessageHandling(t *testing.T) {
	config := TCPServerConfig{
		Address:        ":0",
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		MaxConnections: 5,
	}

	server := NewTCPServer(config)
	defer server.Stop()

	var wg sync.WaitGroup
	var receivedMessage string
	var clientConnection *ClientConnection

	// 设置回调函数
	server.SetCallbacks(
		func(client *ClientConnection) {
			clientConnection = client
		},
		func(client *ClientConnection, err error) {
			// 客户端断开连接
		},
		func(client *ClientConnection, data []byte) {
			receivedMessage = string(data)
			wg.Done()
		},
		func(err error) {
			fmt.Printf("Server error: %v\n", err)
		},
	)

	// 启动服务器
	err := server.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	// 连接到服务器
	conn, err := net.Dial("tcp", server.GetAddress())
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	// 等待连接建立
	time.Sleep(100 * time.Millisecond)

	// 发送消息
	testMessage := "Hello, Server!"
	wg.Add(1)
	_, err = conn.Write([]byte(testMessage))
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	// 等待消息接收
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		if receivedMessage != testMessage {
			t.Errorf("Expected '%s', got '%s'", testMessage, receivedMessage)
		}
	case <-time.After(5 * time.Second):
		t.Error("Timeout waiting for message")
	}

	// 测试服务器向客户端发送消息
	if clientConnection != nil {
		responseMessage := "Hello, Client!"
		err = clientConnection.SendString(responseMessage)
		if err != nil {
			t.Fatalf("Failed to send response: %v", err)
		}

		// 读取响应
		buffer := make([]byte, 1024)
		conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		n, err := conn.Read(buffer)
		if err != nil {
			t.Fatalf("Failed to read response: %v", err)
		}

		response := string(buffer[:n])
		if response != responseMessage {
			t.Errorf("Expected response '%s', got '%s'", responseMessage, response)
		}
	}
}

func TestTCPServer_Broadcast(t *testing.T) {
	config := TCPServerConfig{
		Address:        ":0",
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		MaxConnections: 10,
	}

	server := NewTCPServer(config)
	defer server.Stop()

	// 启动服务器
	err := server.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	// 创建多个客户端连接
	numClients := 3
	var clients []net.Conn
	defer func() {
		for _, conn := range clients {
			conn.Close()
		}
	}()

	for i := 0; i < numClients; i++ {
		conn, err := net.Dial("tcp", server.GetAddress())
		if err != nil {
			t.Fatalf("Failed to connect client %d: %v", i, err)
		}
		clients = append(clients, conn)
	}

	// 等待连接建立
	time.Sleep(200 * time.Millisecond)

	// 检查客户端数量
	if server.GetClientCount() != numClients {
		t.Errorf("Expected %d clients, got %d", numClients, server.GetClientCount())
	}

	// 广播消息
	broadcastMessage := "Broadcast message to all clients!"
	server.BroadcastString(broadcastMessage)

	// 验证所有客户端都收到消息
	for i, conn := range clients {
		buffer := make([]byte, 1024)
		conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		n, err := conn.Read(buffer)
		if err != nil {
			t.Fatalf("Client %d failed to read broadcast message: %v", i, err)
		}

		received := string(buffer[:n])
		if received != broadcastMessage {
			t.Errorf("Client %d: expected '%s', got '%s'", i, broadcastMessage, received)
		}
	}
}

func TestTCPServer_MaxConnections(t *testing.T) {
	maxConnections := 2
	config := TCPServerConfig{
		Address:        ":0",
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		MaxConnections: maxConnections,
	}

	server := NewTCPServer(config)
	defer server.Stop()

	var errorReceived bool
	var wg sync.WaitGroup

	// 设置错误回调
	server.SetCallbacks(
		nil,
		nil,
		nil,
		func(err error) {
			if strings.Contains(err.Error(), "connection limit reached") {
				errorReceived = true
				wg.Done()
			}
		},
	)

	// 启动服务器
	err := server.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	// 创建允许的最大连接数
	var clients []net.Conn
	defer func() {
		for _, conn := range clients {
			conn.Close()
		}
	}()

	for i := 0; i < maxConnections; i++ {
		conn, err := net.Dial("tcp", server.GetAddress())
		if err != nil {
			t.Fatalf("Failed to connect client %d: %v", i, err)
		}
		clients = append(clients, conn)
	}

	// 等待连接建立
	time.Sleep(100 * time.Millisecond)

	// 检查客户端数量
	if server.GetClientCount() != maxConnections {
		t.Errorf("Expected %d clients, got %d", maxConnections, server.GetClientCount())
	}

	// 尝试创建超出限制的连接
	wg.Add(1)
	conn, err := net.Dial("tcp", server.GetAddress())
	if err == nil {
		conn.Close()
	}

	// 等待错误回调
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		if !errorReceived {
			t.Error("Expected connection limit error")
		}
	case <-time.After(5 * time.Second):
		t.Error("Timeout waiting for connection limit error")
	}
}

func TestClientConnection_Methods(t *testing.T) {
	config := TCPServerConfig{
		Address:        ":0",
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		MaxConnections: 5,
	}

	server := NewTCPServer(config)
	defer server.Stop()

	var clientConnection *ClientConnection
	var wg sync.WaitGroup

	// 设置连接回调
	server.SetCallbacks(
		func(client *ClientConnection) {
			clientConnection = client
			wg.Done()
		},
		nil,
		nil,
		nil,
	)

	// 启动服务器
	err := server.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	// 连接到服务器
	wg.Add(1)
	conn, err := net.Dial("tcp", server.GetAddress())
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	// 等待连接建立
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// 测试客户端连接方法
		if clientConnection.ID == "" {
			t.Error("Client ID should not be empty")
		}

		if clientConnection.RemoteAddr == "" {
			t.Error("Client remote address should not be empty")
		}

		if clientConnection.ConnectedAt.IsZero() {
			t.Error("Client connected time should not be zero")
		}

		// 等待一小段时间以确保uptime为正数
		time.Sleep(1 * time.Millisecond)

		uptime := clientConnection.GetUptime()
		if uptime <= 0 {
			t.Error("Client uptime should be positive")
		}

		if clientConnection.IsClosed() {
			t.Error("Client should not be closed initially")
		}

		// 测试关闭连接
		clientConnection.Close()
		if !clientConnection.IsClosed() {
			t.Error("Client should be closed after calling Close()")
		}

	case <-time.After(5 * time.Second):
		t.Error("Timeout waiting for client connection")
	}
}
