package websocket

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestWSServer_StartStop(t *testing.T) {
	config := WSServerConfig{
		Address:        ":0", // 使用随机端口
		Path:           "/ws",
		MaxConnections: 10,
		PingInterval:   5 * time.Second,
		PongWait:       10 * time.Second,
		WriteWait:      5 * time.Second,
	}

	server := NewWSServer(config)

	// 测试启动
	err := server.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	// 检查运行状态
	if !server.IsRunning() {
		t.Error("Server should be running")
	}

	// 等待一段时间让服务器完全启动
	time.Sleep(100 * time.Millisecond)

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

func TestWSClient_ConnectDisconnect(t *testing.T) {
	// 启动WebSocket服务器
	serverConfig := WSServerConfig{
		Address: ":0",
		Path:    "/ws",
	}

	server := NewWSServer(serverConfig)
	defer server.Stop()

	err := server.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	// 等待服务器启动
	time.Sleep(100 * time.Millisecond)

	// 获取服务器地址
	serverAddr := server.GetAddress()
	wsURL := fmt.Sprintf("ws://%s/ws", serverAddr)

	// 创建WebSocket客户端
	clientConfig := WSClientConfig{
		URL:           wsURL,
		AutoReconnect: false,
	}

	client := NewWSClient(clientConfig)
	defer client.Close()

	// 测试连接
	err = client.Connect()
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// 检查连接状态
	if !client.IsConnected() {
		t.Error("Client should be connected")
	}

	// 等待连接建立
	time.Sleep(100 * time.Millisecond)

	// 测试断开连接
	client.Disconnect()

	// 检查连接状态
	if client.IsConnected() {
		t.Error("Client should be disconnected")
	}
}

func TestWSClientServer_MessageExchange(t *testing.T) {
	// 启动WebSocket服务器
	serverConfig := WSServerConfig{
		Address: ":0",
		Path:    "/ws",
	}

	server := NewWSServer(serverConfig)
	defer server.Stop()

	var serverReceivedMessages []string
	var clientReceivedMessages []string

	// 设置服务器回调
	server.SetCallbacks(
		func(client *WSClientConnection) {
			fmt.Printf("Server: Client connected: %s\n", client.RemoteAddr)
			// 向客户端发送欢迎消息
			client.SendText("Welcome to WebSocket server!")
		},
		func(client *WSClientConnection, err error) {
			fmt.Printf("Server: Client disconnected: %s\n", client.RemoteAddr)
		},
		func(client *WSClientConnection, data []byte) {
			message := string(data)
			serverReceivedMessages = append(serverReceivedMessages, message)
			fmt.Printf("Server: Received message: %s\n", message)
			// 回显消息
			client.SendText("Echo: " + message)
		},
		func(err error) {
			fmt.Printf("Server error: %v\n", err)
		},
	)

	err := server.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	// 等待服务器启动
	time.Sleep(100 * time.Millisecond)

	// 获取服务器地址并创建WebSocket URL
	serverAddr := server.GetAddress()
	wsURL := fmt.Sprintf("ws://%s/ws", serverAddr)

	// 创建WebSocket客户端
	clientConfig := WSClientConfig{
		URL:           wsURL,
		AutoReconnect: false,
	}

	client := NewWSClient(clientConfig)
	defer client.Close()

	// 设置客户端回调
	client.SetCallbacks(
		func() {
			fmt.Println("Client: Connected to server")
		},
		func(err error) {
			fmt.Printf("Client: Disconnected from server: %v\n", err)
		},
		func(data []byte) {
			message := string(data)
			clientReceivedMessages = append(clientReceivedMessages, message)
			fmt.Printf("Client: Received message: %s\n", message)
		},
		func(err error) {
			fmt.Printf("Client error: %v\n", err)
		},
	)

	// 连接到服务器
	err = client.Connect()
	if err != nil {
		t.Fatalf("Failed to connect client: %v", err)
	}

	// 等待连接建立和欢迎消息
	time.Sleep(200 * time.Millisecond)

	// 测试消息交换
	testMessages := []string{"Hello", "How are you?", "Goodbye"}

	for _, msg := range testMessages {
		// 客户端发送消息
		err = client.SendText(msg)
		if err != nil {
			t.Fatalf("Failed to send message '%s': %v", msg, err)
		}

		// 等待消息处理
		time.Sleep(100 * time.Millisecond)
	}

	// 验证收到的消息
	if len(serverReceivedMessages) != len(testMessages) {
		t.Errorf("Expected %d messages on server, got %d", len(testMessages), len(serverReceivedMessages))
	}

	// 验证客户端收到的消息（欢迎消息 + 回显消息）
	expectedClientMessages := len(testMessages) + 1 // +1 for welcome message
	if len(clientReceivedMessages) != expectedClientMessages {
		t.Errorf("Expected %d messages on client, got %d", expectedClientMessages, len(clientReceivedMessages))
	}

	// 验证服务器收到的消息内容
	for i, msg := range testMessages {
		if serverReceivedMessages[i] != msg {
			t.Errorf("Expected server message %d to be '%s', got '%s'", i, msg, serverReceivedMessages[i])
		}
	}

	// 验证客户端收到欢迎消息
	if len(clientReceivedMessages) > 0 && clientReceivedMessages[0] != "Welcome to WebSocket server!" {
		t.Errorf("Expected welcome message, got '%s'", clientReceivedMessages[0])
	}
}

func TestWSServer_Broadcast(t *testing.T) {
	// 启动WebSocket服务器
	serverConfig := WSServerConfig{
		Address: ":0",
		Path:    "/ws",
	}

	server := NewWSServer(serverConfig)
	defer server.Stop()

	err := server.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	// 等待服务器启动
	time.Sleep(100 * time.Millisecond)

	// 获取服务器地址
	serverAddr := server.GetAddress()
	wsURL := fmt.Sprintf("ws://%s/ws", serverAddr)

	// 创建多个WebSocket客户端
	numClients := 3
	var clients []*WSClient
	var receivedMessages [][]string = make([][]string, numClients)
	var wg sync.WaitGroup

	for i := 0; i < numClients; i++ {
		clientConfig := WSClientConfig{
			URL:           wsURL,
			AutoReconnect: false,
		}

		client := NewWSClient(clientConfig)
		clients = append(clients, client)

		// 为每个客户端设置回调
		clientIndex := i
		client.SetCallbacks(
			func() {
				fmt.Printf("Client %d: Connected\n", clientIndex)
			},
			func(err error) {
				fmt.Printf("Client %d: Disconnected\n", clientIndex)
			},
			func(data []byte) {
				message := string(data)
				receivedMessages[clientIndex] = append(receivedMessages[clientIndex], message)
				fmt.Printf("Client %d: Received '%s'\n", clientIndex, message)
				wg.Done()
			},
			func(err error) {
				fmt.Printf("Client %d error: %v\n", clientIndex, err)
			},
		)

		// 连接客户端
		err = client.Connect()
		if err != nil {
			t.Fatalf("Failed to connect client %d: %v", i, err)
		}
	}

	// 确保清理客户端
	defer func() {
		for _, client := range clients {
			client.Close()
		}
	}()

	// 等待所有客户端连接
	time.Sleep(300 * time.Millisecond)

	// 检查服务器状态
	if server.GetClientCount() != numClients {
		t.Errorf("Expected %d clients connected, got %d", numClients, server.GetClientCount())
	}

	// 广播消息
	broadcastMessage := "Broadcast message to all clients!"
	wg.Add(numClients) // 每个客户端都应该收到广播消息

	server.BroadcastText(broadcastMessage)

	// 等待消息处理
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// 验证所有客户端都收到了广播消息
		for i := 0; i < numClients; i++ {
			if len(receivedMessages[i]) == 0 {
				t.Errorf("Client %d did not receive any messages", i)
				continue
			}
			// 找到广播消息（可能不是第一条消息，因为可能有其他消息）
			found := false
			for _, msg := range receivedMessages[i] {
				if msg == broadcastMessage {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Client %d did not receive broadcast message", i)
			}
		}
	case <-time.After(5 * time.Second):
		t.Error("Timeout waiting for broadcast messages")
	}
}

func TestWSClient_Reconnect(t *testing.T) {
	// 启动WebSocket服务器（使用固定端口）
	serverConfig := WSServerConfig{
		Address: ":18080", // 使用固定端口避免重连问题
		Path:    "/ws",
	}

	server := NewWSServer(serverConfig)

	err := server.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	// 等待服务器启动
	time.Sleep(100 * time.Millisecond)

	// 获取服务器地址（使用固定地址）
	wsURL := "ws://localhost:18080/ws"

	// 创建支持重连的WebSocket客户端
	clientConfig := WSClientConfig{
		URL:            wsURL,
		AutoReconnect:  true,
		ReconnectDelay: 500 * time.Millisecond,
		MaxReconnects:  3,
	}

	client := NewWSClient(clientConfig)
	defer client.Close()

	var connectCount int
	var wg sync.WaitGroup

	// 设置客户端回调
	client.SetCallbacks(
		func() {
			connectCount++
			fmt.Printf("Client: Connected (count: %d)\n", connectCount)
			wg.Done()
		},
		func(err error) {
			fmt.Printf("Client: Disconnected: %v\n", err)
		},
		nil,
		func(err error) {
			fmt.Printf("Client error: %v\n", err)
		},
	)

	// 第一次连接
	wg.Add(1)
	err = client.Connect()
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// 等待连接建立
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// 连接成功
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for initial connection")
	}

	// 关闭服务器模拟网络断开
	server.Stop()

	// 等待一段时间让客户端检测到断开
	time.Sleep(1 * time.Second)

	// 重新启动服务器
	server = NewWSServer(serverConfig)
	err = server.Start()
	if err != nil {
		t.Fatalf("Failed to restart server: %v", err)
	}
	defer server.Stop()

	// 等待自动重连
	wg.Add(1)
	done2 := make(chan struct{})
	go func() {
		wg.Wait()
		close(done2)
	}()

	select {
	case <-done2:
		if connectCount < 2 {
			t.Errorf("Expected at least 2 connections, got %d", connectCount)
		}
	case <-time.After(5 * time.Second):
		t.Error("Timeout waiting for reconnection")
	}
}
