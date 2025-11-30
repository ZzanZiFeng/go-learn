// Package main demonstrates basic types in Go.
//
// To run this example:
//
//	go run main.go
//
// Expected output:
//
//	=== 整数类型 / Integer Types ===
//	int: 42
//	...
package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

func main() {
	fmt.Println("=== 整数类型 / Integer Types ===")

	// 有符号整数
	// Signed integers
	var i int = 42
	var i8 int8 = 127
	var i16 int16 = 32767
	var i32 int32 = 2147483647
	var i64 int64 = 9223372036854775807

	fmt.Printf("int: %d\n", i)
	fmt.Printf("int8: %d (max: 127)\n", i8)
	fmt.Printf("int16: %d (max: 32767)\n", i16)
	fmt.Printf("int32: %d\n", i32)
	fmt.Printf("int64: %d\n", i64)

	// 无符号整数
	// Unsigned integers
	var u uint = 42
	var u8 uint8 = 255
	var b byte = 65 // byte 是 uint8 的别名

	fmt.Printf("\nuint: %d\n", u)
	fmt.Printf("uint8: %d (max: 255)\n", u8)
	fmt.Printf("byte: %d, as char: %c\n", b, b)

	fmt.Println("\n=== 浮点数类型 / Float Types ===")

	// 浮点数
	// Floating point numbers
	var f32 float32 = 3.14159265358979
	var f64 float64 = 3.14159265358979

	fmt.Printf("float32: %.15f (精度较低)\n", f32)
	fmt.Printf("float64: %.15f (精度较高)\n", f64)
	fmt.Printf("Max float64: %e\n", math.MaxFloat64)

	// 特殊浮点值
	// Special float values
	fmt.Printf("Infinity: %f\n", math.Inf(1))
	fmt.Printf("NaN: %f\n", math.NaN())

	fmt.Println("\n=== 字符串 / Strings ===")

	// 字符串基础
	// String basics
	s1 := "Hello, World!"
	s2 := "你好，世界！"

	fmt.Println("s1:", s1)
	fmt.Println("s2:", s2)
	fmt.Printf("s1 字节数: %d\n", len(s1))
	fmt.Printf("s2 字节数: %d (UTF-8)\n", len(s2))
	fmt.Printf("s2 字符数: %d\n", len([]rune(s2)))

	// 字符串操作
	// String operations
	fmt.Println("\n字符串操作 / String Operations:")
	fmt.Println("ToUpper:", strings.ToUpper(s1))
	fmt.Println("Contains:", strings.Contains(s1, "World"))
	fmt.Println("Split:", strings.Split(s1, ", "))
	fmt.Println("Join:", strings.Join([]string{"a", "b", "c"}, "-"))

	// 原始字符串
	// Raw strings
	rawStr := `Line 1
Line 2
	Indented line`
	fmt.Println("\n原始字符串 / Raw string:")
	fmt.Println(rawStr)

	// 字符串遍历
	// String iteration
	fmt.Println("\n字符串遍历 / String Iteration:")
	sample := "Go语言"
	fmt.Println("按字节遍历:")
	for i := 0; i < len(sample); i++ {
		fmt.Printf("  [%d] byte=%d\n", i, sample[i])
	}
	fmt.Println("按字符遍历 (range):")
	for i, r := range sample {
		fmt.Printf("  [%d] rune=%c (U+%04X)\n", i, r, r)
	}

	fmt.Println("\n=== 布尔类型 / Boolean Type ===")

	// 布尔值
	// Boolean values
	var t bool = true
	var f bool = false

	fmt.Printf("true: %t\n", t)
	fmt.Printf("false: %t\n", f)

	// 比较运算
	// Comparison operations
	x, y := 10, 20
	fmt.Printf("\nx=%d, y=%d\n", x, y)
	fmt.Printf("x == y: %t\n", x == y)
	fmt.Printf("x < y: %t\n", x < y)
	fmt.Printf("x != y: %t\n", x != y)

	// 逻辑运算
	// Logical operations
	fmt.Println("\n逻辑运算 / Logical Operations:")
	fmt.Printf("true && false = %t\n", true && false)
	fmt.Printf("true || false = %t\n", true || false)
	fmt.Printf("!true = %t\n", !true)

	fmt.Println("\n=== 类型转换 / Type Conversion ===")

	// 数值类型转换
	// Numeric type conversion
	intVal := 42
	floatVal := float64(intVal)
	uintVal := uint(floatVal)

	fmt.Printf("int -> float64: %d -> %f\n", intVal, floatVal)
	fmt.Printf("float64 -> uint: %f -> %d\n", floatVal, uintVal)

	// 字符串与数值转换
	// String and number conversion
	numStr := "42"
	num, err := strconv.Atoi(numStr)
	if err == nil {
		fmt.Printf("string -> int: %q -> %d\n", numStr, num)
	}

	str := strconv.Itoa(100)
	fmt.Printf("int -> string: %d -> %q\n", 100, str)

	// 浮点数转换
	// Float conversion
	floatStr := "3.14159"
	floatNum, _ := strconv.ParseFloat(floatStr, 64)
	fmt.Printf("string -> float64: %q -> %f\n", floatStr, floatNum)

	floatToStr := strconv.FormatFloat(3.14159, 'f', 2, 64)
	fmt.Printf("float64 -> string: %f -> %q\n", 3.14159, floatToStr)

	// 使用 fmt.Sprintf
	// Using fmt.Sprintf
	formatted := fmt.Sprintf("value: %d", 42)
	fmt.Printf("fmt.Sprintf: %q\n", formatted)
}
