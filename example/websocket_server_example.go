package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/muchinfo/mtp2-common-lib/websocket"
)

func RunWebSocketServerExample() {
	// 创建WebSocket服务器配置
	config := websocket.WSServerConfig{
		Address:         ":8080",          // 监听端口8080
		Path:            "/ws",            // WebSocket路径
		MaxConnections:  100,              // 最大连接数
		PingInterval:    30 * time.Second, // ping间隔
		PongWait:        60 * time.Second, // pong等待时间
		WriteWait:       10 * time.Second, // 写入等待时间
		ReadBufferSize:  4096,             // 读取缓冲区大小
		WriteBufferSize: 4096,             // 写入缓冲区大小
	}

	// 创建WebSocket服务器
	server := websocket.NewWSServer(config)

	// 设置回调函数
	server.SetCallbacks(
		// 客户端连接回调
		func(client *websocket.WSClientConnection) {
			log.Printf("✅ Client connected: %s from %s", client.ID, client.RemoteAddr)
			log.Printf("🌐 User-Agent: %s", client.UserAgent)

			// 向新连接的客户端发送欢迎消息
			welcomeMsg := map[string]interface{}{
				"type":        "welcome",
				"message":     "Welcome to WebSocket Server! 🎉",
				"server_time": time.Now().Format("2006-01-02 15:04:05"),
				"your_id":     client.ID,
			}

			if err := client.SendJSON(welcomeMsg); err != nil {
				log.Printf("❌ Failed to send welcome message to %s: %v", client.ID, err)
			}

			// 通知其他客户端有新用户加入
			notifyMsg := map[string]interface{}{
				"type":      "user_joined",
				"message":   fmt.Sprintf("User %s joined the chat", client.RemoteAddr),
				"user_id":   client.ID,
				"timestamp": time.Now().Unix(),
			}
			server.BroadcastJSON(notifyMsg)

			log.Printf("📊 Current connections: %d", server.GetClientCount())
		},
		// 客户端断开连接回调
		func(client *websocket.WSClientConnection, err error) {
			if err != nil {
				log.Printf("❌ Client %s disconnected with error: %v", client.ID, err)
			} else {
				log.Printf("📴 Client %s disconnected gracefully", client.ID)
			}

			// 通知其他客户端用户离开
			notifyMsg := map[string]interface{}{
				"type":      "user_left",
				"message":   fmt.Sprintf("User %s left the chat", client.RemoteAddr),
				"user_id":   client.ID,
				"timestamp": time.Now().Unix(),
			}
			server.BroadcastJSON(notifyMsg)

			log.Printf("📊 Current connections: %d", server.GetClientCount())
		},
		// 消息接收回调
		func(client *websocket.WSClientConnection, data []byte) {
			message := strings.TrimSpace(string(data))
			log.Printf("📨 Received from %s: %s", client.RemoteAddr, message)

			// 尝试解析JSON消息
			var jsonMsg map[string]interface{}
			if err := json.Unmarshal(data, &jsonMsg); err == nil {
				handleJSONMessage(client, server, jsonMsg)
				return
			}

			// 处理文本命令
			switch {
			case strings.HasPrefix(message, "/help"):
				handleWSHelpCommand(client)
			case strings.HasPrefix(message, "/list"):
				handleWSListCommand(client, server)
			case strings.HasPrefix(message, "/stats"):
				handleWSStatsCommand(client, server)
			case strings.HasPrefix(message, "/time"):
				handleWSTimeCommand(client)
			case strings.HasPrefix(message, "/echo "):
				handleWSEchoCommand(client, message)
			case strings.HasPrefix(message, "/broadcast "):
				handleWSBroadcastCommand(client, server, message)
			case strings.HasPrefix(message, "/private "):
				handleWSPrivateMessage(client, server, message)
			default:
				// 普通消息，转发给所有客户端
				broadcastMsg := map[string]interface{}{
					"type":      "chat",
					"message":   message,
					"from":      client.RemoteAddr,
					"from_id":   client.ID,
					"timestamp": time.Now().Unix(),
				}
				server.BroadcastJSON(broadcastMsg)
			}
		},
		// 错误回调
		func(err error) {
			log.Printf("⚠️ Server error: %v", err)
		},
	)

	// 启动服务器
	log.Printf("🚀 Starting WebSocket server on %s%s...", config.Address, config.Path)
	if err := server.Start(); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}

	log.Printf("✅ WebSocket server started successfully")
	log.Printf("🌍 WebSocket URL: ws://localhost%s%s", config.Address, config.Path)
	log.Println("📱 Press Ctrl+C to stop the server")

	// 定期打印服务器状态
	go func() {
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			log.Printf("📊 Server status - Address: %s, Path: %s, Connections: %d, Running: %v",
				server.GetAddress(), config.Path, server.GetClientCount(), server.IsRunning())
		}
	}()

	// 等待退出信号
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Println("🛑 Received shutdown signal, stopping server...")

	// 向所有客户端发送服务器关闭通知
	shutdownMsg := map[string]interface{}{
		"type":      "server_shutdown",
		"message":   "Server is shutting down. Goodbye! 👋",
		"timestamp": time.Now().Unix(),
	}
	server.BroadcastJSON(shutdownMsg)
	time.Sleep(1 * time.Second) // 给客户端时间接收消息

	// 停止服务器
	if err := server.Stop(); err != nil {
		log.Printf("❌ Error stopping server: %v", err)
	} else {
		log.Println("✅ Server stopped gracefully")
	}
}

// handleJSONMessage 处理JSON消息
func handleJSONMessage(client *websocket.WSClientConnection, server *websocket.WSServer, jsonMsg map[string]interface{}) {
	msgType, ok := jsonMsg["type"].(string)
	if !ok {
		client.SendJSON(map[string]interface{}{
			"type":    "error",
			"message": "Invalid message format: missing type field",
		})
		return
	}

	switch msgType {
	case "ping":
		// 响应ping消息
		client.SendJSON(map[string]interface{}{
			"type":      "pong",
			"timestamp": time.Now().Unix(),
		})
	case "chat":
		// 聊天消息
		if msg, ok := jsonMsg["message"].(string); ok {
			broadcastMsg := map[string]interface{}{
				"type":      "chat",
				"message":   msg,
				"from":      client.RemoteAddr,
				"from_id":   client.ID,
				"timestamp": time.Now().Unix(),
			}
			server.BroadcastJSON(broadcastMsg)
		}
	default:
		// 转发未知类型的JSON消息
		jsonMsg["from"] = client.RemoteAddr
		jsonMsg["from_id"] = client.ID
		jsonMsg["timestamp"] = time.Now().Unix()
		server.BroadcastJSON(jsonMsg)
	}
}

// handleWSHelpCommand 处理帮助命令
func handleWSHelpCommand(client *websocket.WSClientConnection) {
	helpMsg := map[string]interface{}{
		"type": "help",
		"message": `📚 Available Commands:
/help - Show this help message
/list - List all connected clients
/stats - Show server statistics
/time - Show current server time
/echo <message> - Echo back your message
/broadcast <message> - Broadcast message to all clients
/private <user_id> <message> - Send private message to specific user

💬 Any other message will be broadcasted to all connected clients.
🔧 You can also send JSON messages with different types.`,
	}
	client.SendJSON(helpMsg)
}

// handleWSListCommand 处理列出客户端命令
func handleWSListCommand(client *websocket.WSClientConnection, server *websocket.WSServer) {
	clients := server.GetClients()
	clientList := make([]map[string]interface{}, 0, len(clients))

	for i, c := range clients {
		clientInfo := map[string]interface{}{
			"index":        i + 1,
			"id":           c.ID,
			"remote_addr":  c.RemoteAddr,
			"connected_at": c.ConnectedAt.Format("2006-01-02 15:04:05"),
			"uptime":       c.GetUptime().Round(time.Second).String(),
			"user_agent":   c.UserAgent,
		}
		clientList = append(clientList, clientInfo)
	}

	response := map[string]interface{}{
		"type":    "client_list",
		"message": fmt.Sprintf("👥 Connected clients (%d)", len(clients)),
		"clients": clientList,
	}

	client.SendJSON(response)
}

// handleWSStatsCommand 处理统计命令
func handleWSStatsCommand(client *websocket.WSClientConnection, server *websocket.WSServer) {
	stats := map[string]interface{}{
		"type":    "stats",
		"message": "📊 Server Statistics",
		"data": map[string]interface{}{
			"address":            server.GetAddress(),
			"running":            server.IsRunning(),
			"connected_clients":  server.GetClientCount(),
			"your_connection_id": client.ID,
			"your_uptime":        client.GetUptime().Round(time.Second).String(),
			"server_time":        time.Now().Format("2006-01-02 15:04:05"),
		},
	}

	client.SendJSON(stats)
}

// handleWSTimeCommand 处理时间命令
func handleWSTimeCommand(client *websocket.WSClientConnection) {
	timeMsg := map[string]interface{}{
		"type":      "time",
		"message":   "🕐 Current server time",
		"time":      time.Now().Format("2006-01-02 15:04:05"),
		"timestamp": time.Now().Unix(),
	}
	client.SendJSON(timeMsg)
}

// handleWSEchoCommand 处理回显命令
func handleWSEchoCommand(client *websocket.WSClientConnection, message string) {
	echoMsg := strings.TrimPrefix(message, "/echo ")
	response := map[string]interface{}{
		"type":      "echo",
		"message":   "🔊 Echo",
		"echo":      echoMsg,
		"timestamp": time.Now().Unix(),
	}
	client.SendJSON(response)
}

// handleWSBroadcastCommand 处理广播命令
func handleWSBroadcastCommand(client *websocket.WSClientConnection, server *websocket.WSServer, message string) {
	broadcastMsg := strings.TrimPrefix(message, "/broadcast ")
	fullMsg := map[string]interface{}{
		"type":      "broadcast",
		"message":   "📢 Broadcast message",
		"content":   broadcastMsg,
		"from":      client.RemoteAddr,
		"from_id":   client.ID,
		"timestamp": time.Now().Unix(),
	}
	server.BroadcastJSON(fullMsg)

	// 向发送者确认
	confirmMsg := map[string]interface{}{
		"type":    "confirm",
		"message": "✅ Message broadcasted to all clients",
	}
	client.SendJSON(confirmMsg)
}

// handleWSPrivateMessage 处理私信命令
func handleWSPrivateMessage(client *websocket.WSClientConnection, server *websocket.WSServer, message string) {
	parts := strings.SplitN(strings.TrimPrefix(message, "/private "), " ", 2)
	if len(parts) < 2 {
		client.SendJSON(map[string]interface{}{
			"type":    "error",
			"message": "Usage: /private <user_id> <message>",
		})
		return
	}

	targetID := parts[0]
	privateMsg := parts[1]

	targetClient := server.GetClient(targetID)
	if targetClient == nil {
		client.SendJSON(map[string]interface{}{
			"type":    "error",
			"message": fmt.Sprintf("User %s not found", targetID),
		})
		return
	}

	// 发送私信给目标用户
	msgToTarget := map[string]interface{}{
		"type":      "private_message",
		"message":   "💬 Private message",
		"content":   privateMsg,
		"from":      client.RemoteAddr,
		"from_id":   client.ID,
		"timestamp": time.Now().Unix(),
	}
	targetClient.SendJSON(msgToTarget)

	// 向发送者确认
	client.SendJSON(map[string]interface{}{
		"type":    "confirm",
		"message": fmt.Sprintf("✅ Private message sent to %s", targetClient.RemoteAddr),
	})
}

// 运行说明:
// 1. 运行这个服务器程序: go run server_example.go
// 2. 使用多个客户端连接到服务器进行测试:
//    - 使用websocket_client_example.go
//    - 使用浏览器WebSocket客户端
//    - 使用其他WebSocket客户端工具
// 3. 尝试发送不同的命令和消息
// 4. 按Ctrl+C优雅关闭服务器

/*
输出示例:
2025/07/22 10:30:00 🚀 Starting WebSocket server on :8080/ws...
2025/07/22 10:30:00 ✅ WebSocket server started successfully
2025/07/22 10:30:00 🌍 WebSocket URL: ws://localhost:8080/ws
2025/07/22 10:30:00 📱 Press Ctrl+C to stop the server
2025/07/22 10:30:05 ✅ Client connected: 127.0.0.1:59123_Go-http-client/1.1_1690012205123456789 from 127.0.0.1:59123
2025/07/22 10:30:05 🌐 User-Agent: Go-http-client/1.1
2025/07/22 10:30:05 📊 Current connections: 1
2025/07/22 10:30:05 📨 Received from 127.0.0.1:59123: Hello, WebSocket Server!
2025/07/22 10:30:10 ✅ Client connected: 127.0.0.1:59124_Mozilla/5.0_1690012210987654321 from 127.0.0.1:59124
2025/07/22 10:30:10 📨 Received from 127.0.0.1:59124: /help
2025/07/22 10:31:00 📊 Server status - Address: :8080, Path: /ws, Connections: 2, Running: true
^C2025/07/22 10:31:30 🛑 Received shutdown signal, stopping server...
2025/07/22 10:31:31 ✅ Server stopped gracefully
*/
