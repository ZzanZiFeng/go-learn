// Package main demonstrates a simple Hello World program.
// This is the first Go program you'll write!
//
// To run this example:
//
//	go run main.go
//
// Expected output:
//
//	Hello, World!
//	Hello, Gopher!
//	Welcome to Go programming!
package main

import "fmt"

func main() {
	// 基础输出
	// Basic output using fmt.Println
	fmt.Println("Hello, World!")

	// 使用变量
	// Using variables
	name := "Gopher"
	fmt.Println("Hello,", name+"!")

	// 使用 Printf 格式化输出
	// Formatted output using Printf
	fmt.Printf("Welcome to %s programming!\n", "Go")

	// 多行输出示例
	// Multiple line output example
	fmt.Println()
	fmt.Println("=== Go 语言特点 / Go Language Features ===")
	fmt.Println("1. 简洁 / Simple")
	fmt.Println("2. 高效 / Efficient")
	fmt.Println("3. 并发 / Concurrent")
}
