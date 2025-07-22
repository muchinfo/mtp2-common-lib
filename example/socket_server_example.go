package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/muchinfo/mtp2-common-lib/socket"
)

func RunSocketServerExample() {
	// 创建TCP服务器配置
	config := socket.TCPServerConfig{
		Address:        ":8080",          // 监听端口8080
		ReadTimeout:    30 * time.Second, // 读取超时
		WriteTimeout:   10 * time.Second, // 写入超时
		MaxConnections: 100,              // 最大连接数
	}

	// 创建TCP服务器
	server := socket.NewTCPServer(config)

	// 设置回调函数
	server.SetCallbacks(
		// 客户端连接回调
		func(client *socket.ClientConnection) {
			log.Printf("✅ Client connected: %s from %s", client.ID, client.RemoteAddr)

			// 向新连接的客户端发送欢迎消息
			welcomeMsg := "Welcome to TCP Server! 🎉\n"
			if err := client.SendString(welcomeMsg); err != nil {
				log.Printf("❌ Failed to send welcome message to %s: %v", client.ID, err)
			}

			// 通知其他客户端有新用户加入
			notifyMsg := fmt.Sprintf("📢 User %s joined the chat\n", client.RemoteAddr)
			server.BroadcastString(notifyMsg)
		},
		// 客户端断开连接回调
		func(client *socket.ClientConnection, err error) {
			if err != nil {
				log.Printf("❌ Client %s disconnected with error: %v", client.ID, err)
			} else {
				log.Printf("📴 Client %s disconnected gracefully", client.ID)
			}

			// 通知其他客户端用户离开
			notifyMsg := fmt.Sprintf("📢 User %s left the chat\n", client.RemoteAddr)
			server.BroadcastString(notifyMsg)

			log.Printf("📊 Current connections: %d", server.GetClientCount())
		},
		// 消息接收回调
		func(client *socket.ClientConnection, data []byte) {
			message := strings.TrimSpace(string(data))
			log.Printf("📨 Received from %s: %s", client.RemoteAddr, message)

			// 处理特殊命令
			switch {
			case strings.HasPrefix(message, "/help"):
				handleHelpCommand(client)
			case strings.HasPrefix(message, "/list"):
				handleListCommand(client, server)
			case strings.HasPrefix(message, "/stats"):
				handleStatsCommand(client, server)
			case strings.HasPrefix(message, "/time"):
				handleTimeCommand(client)
			case strings.HasPrefix(message, "/echo "):
				handleEchoCommand(client, message)
			case strings.HasPrefix(message, "/broadcast "):
				handleBroadcastCommand(client, server, message)
			default:
				// 普通消息，转发给所有客户端
				broadcastMsg := fmt.Sprintf("[%s] %s\n", client.RemoteAddr, message)
				server.BroadcastString(broadcastMsg)
			}
		},
		// 错误回调
		func(err error) {
			log.Printf("⚠️ Server error: %v", err)
		},
	)

	// 启动服务器
	log.Printf("🚀 Starting TCP server on %s...", config.Address)
	if err := server.Start(); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}

	log.Printf("✅ TCP server started successfully on %s", server.GetAddress())
	log.Println("📱 Press Ctrl+C to stop the server")

	// 定期打印服务器状态
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			log.Printf("📊 Server status - Address: %s, Connections: %d, Running: %v",
				server.GetAddress(), server.GetClientCount(), server.IsRunning())
		}
	}()

	// 等待退出信号
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Println("🛑 Received shutdown signal, stopping server...")

	// 向所有客户端发送服务器关闭通知
	server.BroadcastString("📢 Server is shutting down. Goodbye! 👋\n")
	time.Sleep(1 * time.Second) // 给客户端时间接收消息

	// 停止服务器
	if err := server.Stop(); err != nil {
		log.Printf("❌ Error stopping server: %v", err)
	} else {
		log.Println("✅ Server stopped gracefully")
	}
}

// handleHelpCommand 处理帮助命令
func handleHelpCommand(client *socket.ClientConnection) {
	helpText := `
📚 Available Commands:
/help - Show this help message
/list - List all connected clients
/stats - Show server statistics
/time - Show current server time
/echo <message> - Echo back your message
/broadcast <message> - Broadcast message to all clients

💬 Any other message will be broadcasted to all connected clients.
`
	client.SendString(helpText)
}

// handleListCommand 处理列出客户端命令
func handleListCommand(client *socket.ClientConnection, server *socket.TCPServer) {
	clients := server.GetClients()
	response := fmt.Sprintf("👥 Connected clients (%d):\n", len(clients))

	for i, c := range clients {
		uptime := c.GetUptime()
		response += fmt.Sprintf("%d. %s (connected %v ago)\n", i+1, c.RemoteAddr, uptime.Round(time.Second))
	}

	client.SendString(response)
}

// handleStatsCommand 处理统计命令
func handleStatsCommand(client *socket.ClientConnection, server *socket.TCPServer) {
	stats := fmt.Sprintf(`
📊 Server Statistics:
- Address: %s
- Running: %v
- Connected clients: %d
- Your connection ID: %s
- Your uptime: %v
`,
		server.GetAddress(),
		server.IsRunning(),
		server.GetClientCount(),
		client.ID,
		client.GetUptime().Round(time.Second),
	)

	client.SendString(stats)
}

// handleTimeCommand 处理时间命令
func handleTimeCommand(client *socket.ClientConnection) {
	timeStr := fmt.Sprintf("🕐 Current server time: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	client.SendString(timeStr)
}

// handleEchoCommand 处理回显命令
func handleEchoCommand(client *socket.ClientConnection, message string) {
	echoMsg := strings.TrimPrefix(message, "/echo ")
	response := fmt.Sprintf("🔊 Echo: %s\n", echoMsg)
	client.SendString(response)
}

// handleBroadcastCommand 处理广播命令
func handleBroadcastCommand(client *socket.ClientConnection, server *socket.TCPServer, message string) {
	broadcastMsg := strings.TrimPrefix(message, "/broadcast ")
	fullMsg := fmt.Sprintf("📢 Broadcast from %s: %s\n", client.RemoteAddr, broadcastMsg)
	server.BroadcastString(fullMsg)

	// 向发送者确认
	client.SendString("✅ Message broadcasted to all clients\n")
}

// 运行说明:
// 1. 运行这个服务器程序: go run server_example.go
// 2. 使用多个终端连接到服务器进行测试:
//    - telnet localhost 8080
//    - nc localhost 8080
//    - 或者使用之前的TCP客户端
// 3. 尝试发送不同的命令和消息
// 4. 按Ctrl+C优雅关闭服务器

/*
输出示例:
2025/07/22 10:30:00 🚀 Starting TCP server on :8080...
2025/07/22 10:30:00 ✅ TCP server started successfully on [::]:8080
2025/07/22 10:30:00 📱 Press Ctrl+C to stop the server
2025/07/22 10:30:05 ✅ Client connected: [::1]:59123_1690012205123456789 from [::1]:59123
2025/07/22 10:30:05 📨 Received from [::1]:59123: Hello, Server!
2025/07/22 10:30:10 ✅ Client connected: [::1]:59124_1690012210987654321 from [::1]:59124
2025/07/22 10:30:10 📨 Received from [::1]:59124: /help
2025/07/22 10:30:30 📊 Server status - Address: [::]:8080, Connections: 2, Running: true
^C2025/07/22 10:31:00 🛑 Received shutdown signal, stopping server...
2025/07/22 10:31:01 ✅ Server stopped gracefully
*/
