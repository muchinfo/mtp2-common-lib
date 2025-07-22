package socket

import (
	"fmt"
	"net"
	"sync"
	"testing"
	"time"
)

// 创建一个简单的TCP服务器用于测试
func startTestServer(port string) (net.Listener, error) {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}

			// 处理连接
			go func(c net.Conn) {
				defer c.Close()
				buffer := make([]byte, 1024)
				for {
					n, err := c.Read(buffer)
					if err != nil {
						return
					}
					// 回显收到的数据
					c.Write(buffer[:n])
				}
			}(conn)
		}
	}()

	return listener, nil
}

func TestTCPClient_Connect(t *testing.T) {
	// 启动测试服务器
	listener, err := startTestServer("8080")
	if err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}
	defer listener.Close()

	// 给服务器一点时间启动
	time.Sleep(100 * time.Millisecond)

	// 创建TCP客户端
	config := TCPClientConfig{
		Address:        "localhost:8080",
		ReconnectDelay: 2 * time.Second,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		MaxReconnects:  3,
		AutoReconnect:  true,
	}

	client := NewTCPClient(config)
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

	// 测试发送数据
	testMessage := "Hello, Server!"
	err = client.SendString(testMessage)
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	// 断开连接
	client.Disconnect()

	// 检查连接状态
	if client.IsConnected() {
		t.Error("Client should be disconnected")
	}
}

func TestTCPClient_Callbacks(t *testing.T) {
	// 启动测试服务器
	listener, err := startTestServer("8081")
	if err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}
	defer listener.Close()

	time.Sleep(100 * time.Millisecond)

	config := TCPClientConfig{
		Address:        "localhost:8081",
		ReconnectDelay: 1 * time.Second,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		MaxReconnects:  2,
		AutoReconnect:  false,
	}

	client := NewTCPClient(config)
	defer client.Close()

	var wg sync.WaitGroup
	var receivedMessage string

	// 设置回调函数
	client.SetCallbacks(
		func() {
			fmt.Println("Connected to server")
		},
		func(err error) {
			fmt.Printf("Disconnected from server: %v\n", err)
		},
		func(data []byte) {
			receivedMessage = string(data)
			wg.Done()
		},
		func(err error) {
			fmt.Printf("Error occurred: %v\n", err)
		},
	)

	// 连接并发送消息
	err = client.Connect()
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	testMessage := "Test Message"
	wg.Add(1)

	err = client.SendString(testMessage)
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	// 等待接收到回显消息
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
}

func TestTCPClient_AutoReconnect(t *testing.T) {
	// 启动测试服务器
	listener, err := startTestServer("8082")
	if err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	config := TCPClientConfig{
		Address:        "localhost:8082",
		ReconnectDelay: 500 * time.Millisecond,
		ReadTimeout:    2 * time.Second,
		WriteTimeout:   2 * time.Second,
		MaxReconnects:  2,
		AutoReconnect:  true,
	}

	client := NewTCPClient(config)
	defer client.Close()

	var reconnectCount int
	var wg sync.WaitGroup

	// 设置回调函数
	client.SetCallbacks(
		func() {
			reconnectCount++
			wg.Done()
		},
		func(err error) {
			fmt.Printf("Disconnected: %v\n", err)
		},
		nil,
		func(err error) {
			fmt.Printf("Error: %v\n", err)
		},
	)

	// 第一次连接
	wg.Add(1)
	err = client.Connect()
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// 等待连接成功
	wg.Wait()

	// 关闭服务器模拟网络断开
	listener.Close()

	// 等待一段时间让客户端检测到断开
	time.Sleep(1 * time.Second)

	// 重新启动服务器
	listener, err = startTestServer("8082")
	if err != nil {
		t.Fatalf("Failed to restart test server: %v", err)
	}
	defer listener.Close()

	// 等待自动重连
	wg.Add(1)
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		if reconnectCount < 2 {
			t.Errorf("Expected at least 2 connections, got %d", reconnectCount)
		}
	case <-time.After(5 * time.Second):
		t.Error("Timeout waiting for reconnection")
	}
}
