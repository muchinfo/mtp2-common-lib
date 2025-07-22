package socket

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// TestTCPClientServerIntegration 集成测试：客户端和服务器协同工作
func TestTCPClientServerIntegration(t *testing.T) {
	// 创建服务器
	serverConfig := TCPServerConfig{
		Address:        ":0", // 使用随机端口
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		MaxConnections: 10,
	}

	server := NewTCPServer(serverConfig)
	defer server.Stop()

	var wg sync.WaitGroup
	var receivedMessages []string
	var connectedClients int

	// 设置服务器回调
	server.SetCallbacks(
		func(client *ClientConnection) {
			connectedClients++
			fmt.Printf("Server: Client %s connected\n", client.RemoteAddr)
			// 向新客户端发送欢迎消息
			client.SendString("Welcome to the server!")
		},
		func(client *ClientConnection, err error) {
			connectedClients--
			fmt.Printf("Server: Client %s disconnected\n", client.RemoteAddr)
		},
		func(client *ClientConnection, data []byte) {
			message := string(data)
			receivedMessages = append(receivedMessages, message)
			fmt.Printf("Server: Received '%s' from %s\n", message, client.RemoteAddr)

			// 回显消息给客户端
			client.SendString("Echo: " + message)
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

	// 创建客户端
	clientConfig := TCPClientConfig{
		Address:        server.GetAddress(),
		ReconnectDelay: 2 * time.Second,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		MaxReconnects:  3,
		AutoReconnect:  false,
	}

	client := NewTCPClient(clientConfig)
	defer client.Close()

	var clientReceivedMessages []string

	// 设置客户端回调
	client.SetCallbacks(
		func() {
			fmt.Printf("Client: Connected to server %s\n", client.GetAddress())
		},
		func(err error) {
			fmt.Printf("Client: Disconnected from server: %v\n", err)
		},
		func(data []byte) {
			message := string(data)
			clientReceivedMessages = append(clientReceivedMessages, message)
			fmt.Printf("Client: Received '%s'\n", message)
			// 只有非欢迎消息才调用Done
			if message != "Welcome to the server!" {
				wg.Done()
			}
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

	// 等待连接建立
	time.Sleep(100 * time.Millisecond)

	// 检查连接状态
	if !client.IsConnected() {
		t.Error("Client should be connected")
	}

	if !server.IsRunning() {
		t.Error("Server should be running")
	}

	if server.GetClientCount() != 1 {
		t.Errorf("Expected 1 client connected, got %d", server.GetClientCount())
	}

	// 测试消息交换
	testMessages := []string{"Hello", "How are you?", "Goodbye"}

	for _, msg := range testMessages {
		wg.Add(2) // 服务器接收 + 客户端接收回显

		// 客户端发送消息
		err = client.SendString(msg)
		if err != nil {
			t.Fatalf("Failed to send message '%s': %v", msg, err)
		}

		// 等待消息处理完成
		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			// 消息处理完成
		case <-time.After(5 * time.Second):
			t.Fatalf("Timeout waiting for message '%s' to be processed", msg)
		}
	}

	// 验证收到的消息
	if len(receivedMessages) != len(testMessages) {
		t.Errorf("Expected %d messages on server, got %d", len(testMessages), len(receivedMessages))
	}

	// 验证客户端收到的消息（欢迎消息 + 回显消息）
	expectedClientMessages := len(testMessages) + 1 // +1 for welcome message
	if len(clientReceivedMessages) != expectedClientMessages {
		t.Errorf("Expected %d messages on client, got %d", expectedClientMessages, len(clientReceivedMessages))
	}

	// 验证消息内容
	for i, msg := range testMessages {
		if receivedMessages[i] != msg {
			t.Errorf("Expected server message %d to be '%s', got '%s'", i, msg, receivedMessages[i])
		}
	}

	// 验证客户端收到欢迎消息
	if len(clientReceivedMessages) > 0 && clientReceivedMessages[0] != "Welcome to the server!" {
		t.Errorf("Expected welcome message, got '%s'", clientReceivedMessages[0])
	}
}

// TestTCPServerMultipleClients 测试服务器处理多个客户端
func TestTCPServerMultipleClients(t *testing.T) {
	// 创建服务器
	serverConfig := TCPServerConfig{
		Address:        ":0",
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		MaxConnections: 5,
	}

	server := NewTCPServer(serverConfig)
	defer server.Stop()

	var wg sync.WaitGroup
	var connectedClients int
	var mu sync.Mutex

	// 设置服务器回调
	server.SetCallbacks(
		func(client *ClientConnection) {
			mu.Lock()
			connectedClients++
			mu.Unlock()
			fmt.Printf("Server: Client %s connected (total: %d)\n", client.RemoteAddr, connectedClients)
		},
		func(client *ClientConnection, err error) {
			mu.Lock()
			connectedClients--
			mu.Unlock()
			fmt.Printf("Server: Client %s disconnected (total: %d)\n", client.RemoteAddr, connectedClients)
		},
		func(client *ClientConnection, data []byte) {
			message := string(data)
			fmt.Printf("Server: Broadcasting message '%s'\n", message)
			// 广播消息给所有客户端
			server.BroadcastString("Broadcast: " + message)
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

	// 创建多个客户端
	numClients := 3
	var clients []*TCPClient
	var receivedMessages [][]string = make([][]string, numClients)

	for i := 0; i < numClients; i++ {
		clientConfig := TCPClientConfig{
			Address:       server.GetAddress(),
			ReadTimeout:   5 * time.Second,
			WriteTimeout:  5 * time.Second,
			AutoReconnect: false,
		}

		client := NewTCPClient(clientConfig)
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
	time.Sleep(200 * time.Millisecond)

	// 检查服务器状态
	if server.GetClientCount() != numClients {
		t.Errorf("Expected %d clients connected, got %d", numClients, server.GetClientCount())
	}

	// 从第一个客户端发送消息
	testMessage := "Hello from client 0"
	wg.Add(1 + numClients) // 1 for server receive + numClients for broadcast

	err = clients[0].SendString(testMessage)
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	// 等待消息处理
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// 验证所有客户端都收到了广播消息
		expectedBroadcast := "Broadcast: " + testMessage
		for i := 0; i < numClients; i++ {
			if len(receivedMessages[i]) == 0 {
				t.Errorf("Client %d did not receive any messages", i)
				continue
			}
			if receivedMessages[i][0] != expectedBroadcast {
				t.Errorf("Client %d expected '%s', got '%s'", i, expectedBroadcast, receivedMessages[i][0])
			}
		}
	case <-time.After(5 * time.Second):
		t.Error("Timeout waiting for broadcast messages")
	}
}
