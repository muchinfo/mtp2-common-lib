package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/muchinfo/mtp2-common-lib/websocket"
)

func RunWebSocketClientExample() {
	// 创建WebSocket客户端配置
	config := websocket.WSClientConfig{
		URL:            "ws://localhost:8080/ws", // WebSocket服务器地址
		ReconnectDelay: 5 * time.Second,          // 重连延迟
		MaxReconnects:  0,                        // 0表示无限重连
		AutoReconnect:  true,                     // 启用自动重连
		PingInterval:   30 * time.Second,         // ping间隔
		PongWait:       60 * time.Second,         // pong等待时间
		WriteWait:      10 * time.Second,         // 写入等待时间
	}

	// 创建WebSocket客户端
	client := websocket.NewWSClient(config)
	defer client.Close()

	// 设置回调函数
	client.SetCallbacks(
		// 连接成功回调
		func() {
			log.Printf("✅ Successfully connected to %s", client.GetURL())
			log.Printf("🔄 Reconnect count: %d", client.GetReconnectCount())
		},
		// 断开连接回调
		func(err error) {
			if err != nil {
				log.Printf("❌ Disconnected from server: %v", err)
			} else {
				log.Printf("📴 Disconnected from server")
			}
		},
		// 消息接收回调
		func(data []byte) {
			log.Printf("📨 Received message: %s", string(data))
		},
		// 错误回调
		func(err error) {
			log.Printf("⚠️ Error occurred: %v", err)
		},
	)

	// 连接到WebSocket服务器
	log.Printf("🔗 Connecting to %s...", config.URL)
	if err := client.Connect(); err != nil {
		log.Fatalf("❌ Failed to connect: %v", err)
	}

	// 等待连接建立
	time.Sleep(1 * time.Second)

	// 发送一些测试消息
	messages := []string{
		"Hello, WebSocket Server!",
		"This is a test message",
		"WebSocket Client is working!",
		"JSON message example",
		"Goodbye!",
	}

	for i, msg := range messages {
		if !client.IsConnected() {
			log.Printf("⚠️ Client is not connected, skipping message %d", i+1)
			continue
		}

		log.Printf("📤 Sending message %d: %s", i+1, msg)

		// 发送文本消息
		if err := client.SendText(msg); err != nil {
			log.Printf("❌ Failed to send text message %d: %v", i+1, err)
		}

		// 等待一段时间再发送下一条消息
		time.Sleep(2 * time.Second)

		// 每三条消息发送一个JSON消息
		if (i+1)%3 == 0 {
			jsonMsg := map[string]interface{}{
				"type":      "json",
				"message":   fmt.Sprintf("JSON message #%d", i+1),
				"timestamp": time.Now().Unix(),
			}

			log.Printf("📤 Sending JSON message: %+v", jsonMsg)
			if err := client.SendJSON(jsonMsg); err != nil {
				log.Printf("❌ Failed to send JSON message: %v", err)
			}

			time.Sleep(2 * time.Second)
		}
	}

	// 等待用户按Ctrl+C退出
	log.Println("📱 Press Ctrl+C to exit...")

	// 设置信号处理
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// 等待信号
	<-c
	log.Println("🛑 Received shutdown signal, closing client...")

	// 客户端会在defer中自动关闭
}

// 运行说明:
// 1. 首先启动WebSocket服务器（使用server_example.go或其他WebSocket服务器）
// 2. 然后运行这个客户端程序: go run client_example.go
// 3. 客户端会自动连接到服务器并发送消息
// 4. 如果连接断开，客户端会自动尝试重连
// 5. 按Ctrl+C优雅关闭客户端

/*
输出示例:
2025/07/22 10:30:00 🔗 Connecting to ws://localhost:8080/ws...
2025/07/22 10:30:00 ✅ Successfully connected to ws://localhost:8080/ws
2025/07/22 10:30:00 🔄 Reconnect count: 0
2025/07/22 10:30:01 📤 Sending message 1: Hello, WebSocket Server!
2025/07/22 10:30:01 📨 Received message: Echo: Hello, WebSocket Server!
2025/07/22 10:30:03 📤 Sending message 2: This is a test message
2025/07/22 10:30:03 📨 Received message: Echo: This is a test message
2025/07/22 10:30:05 📤 Sending message 3: WebSocket Client is working!
2025/07/22 10:30:05 📨 Received message: Echo: WebSocket Client is working!
2025/07/22 10:30:07 📤 Sending JSON message: map[message:JSON message #3 timestamp:1690012207 type:json]
2025/07/22 10:30:07 📨 Received message: {"type":"json","message":"JSON message #3","timestamp":1690012207}
2025/07/22 10:30:09 📱 Press Ctrl+C to exit...
^C2025/07/22 10:30:15 🛑 Received shutdown signal, closing client...
*/
