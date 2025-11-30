// Package main demonstrates variable and constant declarations in Go.
//
// To run this example:
//
//	go run main.go
//
// Expected output:
//
//	=== 变量声明 / Variable Declarations ===
//	Name: Gopher
//	Age: 25
//	...
package main

import "fmt"

func main() {
	fmt.Println("=== 变量声明 / Variable Declarations ===")

	// 使用 var 声明变量
	// Declare variables with var
	var name string = "Gopher"
	var age int = 25
	var isActive bool = true

	fmt.Println("Name:", name)
	fmt.Println("Age:", age)
	fmt.Println("Is Active:", isActive)

	// 类型推断
	// Type inference
	var city = "Beijing"
	fmt.Println("City:", city)

	// 多变量声明
	// Multiple variable declaration
	var (
		firstName = "John"
		lastName  = "Doe"
		score     = 95.5
	)
	fmt.Printf("Full name: %s %s, Score: %.1f\n", firstName, lastName, score)

	fmt.Println("\n=== 短变量声明 / Short Variable Declaration ===")

	// 短变量声明 (:=)
	// Short variable declaration (:=)
	language := "Go"
	version := 1.21
	fmt.Printf("Language: %s, Version: %.2f\n", language, version)

	// 同时声明多个变量
	// Declare multiple variables at once
	x, y, z := 10, 20, 30
	fmt.Printf("x=%d, y=%d, z=%d\n", x, y, z)

	// 交换变量值
	// Swap variable values
	x, y = y, x
	fmt.Printf("After swap: x=%d, y=%d\n", x, y)

	fmt.Println("\n=== 常量 / Constants ===")

	// 常量声明
	// Constant declaration
	const Pi = 3.14159
	const Greeting = "Hello"
	fmt.Println("Pi:", Pi)
	fmt.Println("Greeting:", Greeting)

	// 常量组
	// Constant group
	const (
		StatusOK       = 200
		StatusNotFound = 404
		StatusError    = 500
	)
	fmt.Printf("HTTP Status: OK=%d, NotFound=%d, Error=%d\n",
		StatusOK, StatusNotFound, StatusError)

	fmt.Println("\n=== iota 枚举 / iota Enumeration ===")

	// 使用 iota 生成枚举值
	// Using iota for enumeration
	const (
		Sunday = iota // 0
		Monday        // 1
		Tuesday       // 2
		Wednesday     // 3
		Thursday      // 4
		Friday        // 5
		Saturday      // 6
	)
	fmt.Printf("Sunday=%d, Monday=%d, Friday=%d\n", Sunday, Monday, Friday)

	// iota 用于位掩码
	// iota for bitmask
	const (
		ReadPerm  = 1 << iota // 1 << 0 = 1
		WritePerm             // 1 << 1 = 2
		ExecPerm              // 1 << 2 = 4
	)
	fmt.Printf("Permissions: Read=%d, Write=%d, Exec=%d\n",
		ReadPerm, WritePerm, ExecPerm)

	// iota 用于存储单位
	// iota for storage units
	const (
		_  = iota             // 忽略 0
		KB = 1 << (10 * iota) // 1 << 10
		MB                    // 1 << 20
		GB                    // 1 << 30
	)
	fmt.Printf("Storage: KB=%d, MB=%d, GB=%d\n", KB, MB, GB)

	fmt.Println("\n=== 零值 / Zero Values ===")

	// 零值示例
	// Zero value examples
	var i int
	var f float64
	var b bool
	var s string
	var p *int

	fmt.Printf("int 零值: %d\n", i)
	fmt.Printf("float64 零值: %f\n", f)
	fmt.Printf("bool 零值: %t\n", b)
	fmt.Printf("string 零值: %q\n", s)
	fmt.Printf("pointer 零值: %v\n", p)

	fmt.Println("\n=== 类型检查 / Type Check ===")

	// 使用 %T 检查类型
	// Use %T to check type
	auto1 := 42
	auto2 := 3.14
	auto3 := "hello"
	auto4 := true

	fmt.Printf("auto1: type=%T, value=%v\n", auto1, auto1)
	fmt.Printf("auto2: type=%T, value=%v\n", auto2, auto2)
	fmt.Printf("auto3: type=%T, value=%v\n", auto3, auto3)
	fmt.Printf("auto4: type=%T, value=%v\n", auto4, auto4)
}
