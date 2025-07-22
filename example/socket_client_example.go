package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/muchinfo/mtp2-common-lib/socket"
)

func RunSocketClientExample() {
	// 创建TCP客户端配置
	config := socket.TCPClientConfig{
		Address:        "localhost:8080", // 服务器地址
		ReconnectDelay: 5 * time.Second,  // 重连延迟
		ReadTimeout:    30 * time.Second, // 读取超时
		WriteTimeout:   10 * time.Second, // 写入超时
		MaxReconnects:  0,                // 0表示无限重连
		AutoReconnect:  true,             // 自动重连
	}

	// 创建TCP客户端
	client := socket.NewTCPClient(config)
	defer client.Close()

	// 设置回调函数
	client.SetCallbacks(
		// 连接成功回调
		func() {
			log.Printf("✅ Successfully connected to %s", client.GetAddress())
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

	// 连接到服务器
	log.Printf("🔗 Connecting to %s...", config.Address)
	if err := client.Connect(); err != nil {
		log.Fatalf("❌ Failed to connect: %v", err)
	}

	// 等待连接建立
	time.Sleep(1 * time.Second)

	// 发送一些测试消息
	messages := []string{
		"Hello, Server!",
		"This is a test message",
		"TCP Client is working!",
		"Goodbye!",
	}

	for i, msg := range messages {
		if !client.IsConnected() {
			log.Printf("⚠️ Client is not connected, skipping message %d", i+1)
			continue
		}

		log.Printf("📤 Sending message %d: %s", i+1, msg)
		if err := client.SendString(msg + "\n"); err != nil {
			log.Printf("❌ Failed to send message %d: %v", i+1, err)
		}

		// 等待一段时间再发送下一条消息
		time.Sleep(2 * time.Second)
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
// 1. 首先启动一个TCP服务器（可以使用nc命令: nc -l -p 8080）
// 2. 然后运行这个客户端程序: go run main.go
// 3. 客户端会自动连接到服务器并发送消息
// 4. 如果连接断开，客户端会自动尝试重连

/*
输出示例:
2025/07/22 10:30:00 🔗 Connecting to localhost:8080...
2025/07/22 10:30:00 ✅ Successfully connected to localhost:8080
2025/07/22 10:30:01 📤 Sending message 1: Hello, Server!
2025/07/22 10:30:01 📨 Received message: Hello, Server!
2025/07/22 10:30:03 📤 Sending message 2: This is a test message
2025/07/22 10:30:03 📨 Received message: This is a test message
2025/07/22 10:30:05 📤 Sending message 3: TCP Client is working!
2025/07/22 10:30:05 📨 Received message: TCP Client is working!
2025/07/22 10:30:07 📤 Sending message 4: Goodbye!
2025/07/22 10:30:07 📨 Received message: Goodbye!
2025/07/22 10:30:09 📱 Press Ctrl+C to exit...
^C2025/07/22 10:30:15 🛑 Received shutdown signal, closing client...
*/
