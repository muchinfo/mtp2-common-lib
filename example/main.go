package main

import "fmt"

func main() {
	fmt.Println("--- Logger Example ---")
	RunLoggerExample()

	fmt.Println("--- Viper Example ---")
	RunViperExample()

	fmt.Println("--- RabbitMQ Example ---")
	RunRabbitMQExample()

	fmt.Println("--- Xorm Oracle Example ---")
	RunXormOracleExample()
}
