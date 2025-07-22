package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("用法: go run . [模块名]")
		fmt.Println("可选模块: viper | logger | rabbitmq | xorm | ulidgen | http | socket_client | socket_server | all")
		return
	}
	arg := strings.ToLower(os.Args[1])
	switch arg {
	case "viper":
		fmt.Println("--- Viper Example ---")
		RunViperExample()
	case "logger":
		fmt.Println("--- Logger Example ---")
		RunLoggerExample()
	case "rabbitmq":
		fmt.Println("--- RabbitMQ Example ---")
		RunRabbitMQExample()
	case "xorm":
		fmt.Println("--- Xorm Oracle Example ---")
		RunXormOracleExample()
	case "ulidgen":
		fmt.Println("--- ULID Gen Example ---")
		RunULIDGenExample()
	case "http":
		fmt.Println("--- HTTP Example ---")
		RunHttpExample()
	case "socket_client":
		fmt.Println("--- Socket Client Example ---")
		RunSocketClientExample()
	case "socket_server":
		fmt.Println("--- Socket Server Example ---")
		RunSocketServerExample()
	case "all":
		fmt.Println("--- Viper Example ---")
		RunViperExample()
		fmt.Println("--- Logger Example ---")
		RunLoggerExample()
		fmt.Println("--- RabbitMQ Example ---")
		RunRabbitMQExample()
		fmt.Println("--- Xorm Oracle Example ---")
		RunXormOracleExample()
		fmt.Println("--- ULID Gen Example ---")
		RunULIDGenExample()
		fmt.Println("--- HTTP Example ---")
		RunHttpExample()
		fmt.Println("--- Socket Client Example ---")
		RunSocketClientExample()
		fmt.Println("--- Socket Server Example ---")
		RunSocketServerExample()
	default:
		fmt.Println("未知模块:", arg)
		fmt.Println("可选模块: viper | logger | rabbitmq | xorm | ulidgen | http | socket_client | socket_server | all")
	}
}
