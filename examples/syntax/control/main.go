// Package main demonstrates control flow in Go.
//
// To run this example:
//
//	go run main.go
//
// Expected output:
//
//	=== if 语句 / if Statements ===
//	x 大于 5
//	...
package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("=== if 语句 / if Statements ===")

	x := 10

	// 基本 if
	// Basic if
	if x > 5 {
		fmt.Println("x 大于 5")
	}

	// if-else
	if x > 15 {
		fmt.Println("x 大于 15")
	} else {
		fmt.Println("x 不大于 15")
	}

	// if-else if-else
	if x < 5 {
		fmt.Println("x 小于 5")
	} else if x < 10 {
		fmt.Println("x 小于 10")
	} else {
		fmt.Println("x 大于等于 10")
	}

	// 带初始化语句的 if
	// if with initialization
	if num := 20; num > 10 {
		fmt.Println("num 大于 10:", num)
	}
	// num 在这里不可用

	fmt.Println("\n=== for 循环 / for Loops ===")

	// 传统 for 循环
	// Traditional for loop
	fmt.Print("传统 for: ")
	for i := 0; i < 5; i++ {
		fmt.Print(i, " ")
	}
	fmt.Println()

	// while 风格
	// while style
	fmt.Print("while 风格: ")
	j := 0
	for j < 5 {
		fmt.Print(j, " ")
		j++
	}
	fmt.Println()

	// range 遍历切片
	// range over slice
	nums := []int{10, 20, 30, 40, 50}
	fmt.Println("range 遍历切片:")
	for i, v := range nums {
		fmt.Printf("  index=%d, value=%d\n", i, v)
	}

	// range 遍历 map
	// range over map
	colors := map[string]string{"red": "#FF0000", "green": "#00FF00"}
	fmt.Println("range 遍历 map:")
	for k, v := range colors {
		fmt.Printf("  %s: %s\n", k, v)
	}

	// range 遍历字符串
	// range over string
	fmt.Println("range 遍历字符串 'Go语言':")
	for i, r := range "Go语言" {
		fmt.Printf("  [%d] %c\n", i, r)
	}

	fmt.Println("\n=== break 和 continue ===")

	// break
	fmt.Print("break 示例: ")
	for i := 0; i < 10; i++ {
		if i == 5 {
			break
		}
		fmt.Print(i, " ")
	}
	fmt.Println()

	// continue
	fmt.Print("continue 示例 (奇数): ")
	for i := 0; i < 10; i++ {
		if i%2 == 0 {
			continue
		}
		fmt.Print(i, " ")
	}
	fmt.Println()

	// 带标签的 break
	// Labeled break
	fmt.Println("带标签的 break:")
outer:
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if i == 1 && j == 1 {
				fmt.Println("  跳出外层循环")
				break outer
			}
			fmt.Printf("  (%d, %d)\n", i, j)
		}
	}

	fmt.Println("\n=== switch 语句 / switch Statements ===")

	// 基本 switch
	// Basic switch
	day := 3
	fmt.Print("基本 switch: ")
	switch day {
	case 1:
		fmt.Println("Monday")
	case 2:
		fmt.Println("Tuesday")
	case 3:
		fmt.Println("Wednesday")
	case 4:
		fmt.Println("Thursday")
	case 5:
		fmt.Println("Friday")
	case 6, 7:
		fmt.Println("Weekend")
	default:
		fmt.Println("Invalid day")
	}

	// 无条件 switch
	// Condition-less switch
	hour := time.Now().Hour()
	fmt.Print("无条件 switch: ")
	switch {
	case hour < 12:
		fmt.Println("Good morning!")
	case hour < 17:
		fmt.Println("Good afternoon!")
	case hour < 21:
		fmt.Println("Good evening!")
	default:
		fmt.Println("Good night!")
	}

	// 带初始化的 switch
	// switch with initialization
	switch score := 85; {
	case score >= 90:
		fmt.Println("Grade: A")
	case score >= 80:
		fmt.Println("Grade: B")
	case score >= 70:
		fmt.Println("Grade: C")
	default:
		fmt.Println("Grade: F")
	}

	// fallthrough
	fmt.Println("\nfallthrough 示例:")
	n := 2
	switch n {
	case 1:
		fmt.Println("  One")
		fallthrough
	case 2:
		fmt.Println("  Two")
		fallthrough
	case 3:
		fmt.Println("  Three")
	case 4:
		fmt.Println("  Four")
	}

	// 类型 switch
	// Type switch
	fmt.Println("\n类型 switch:")
	checkType(42)
	checkType("hello")
	checkType(true)
	checkType(3.14)
}

func checkType(i interface{}) {
	switch v := i.(type) {
	case int:
		fmt.Printf("  int: %d\n", v)
	case string:
		fmt.Printf("  string: %s\n", v)
	case bool:
		fmt.Printf("  bool: %t\n", v)
	default:
		fmt.Printf("  unknown type: %T\n", v)
	}
}
