// Package main demonstrates functions in Go.
//
// To run this example:
//
//	go run main.go
//
// Expected output:
//
//	=== 基本函数 / Basic Functions ===
//	add(3, 5) = 8
//	...
package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("=== 基本函数 / Basic Functions ===")

	// 调用简单函数
	// Call simple functions
	fmt.Println("add(3, 5) =", add(3, 5))
	fmt.Println("multiply(4, 6) =", multiply(4, 6))

	sayHello()
	greet("Gopher")

	fmt.Println("\n=== 多返回值 / Multiple Return Values ===")

	// 多返回值
	// Multiple return values
	sum, product := sumAndProduct(3, 4)
	fmt.Printf("sumAndProduct(3, 4) = sum: %d, product: %d\n", sum, product)

	// 带错误返回
	// With error return
	result, err := divide(10, 2)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("divide(10, 2) =", result)
	}

	result, err = divide(10, 0)
	if err != nil {
		fmt.Println("divide(10, 0) Error:", err)
	}

	fmt.Println("\n=== 命名返回值 / Named Return Values ===")

	area, perimeter := rectangle(5, 3)
	fmt.Printf("rectangle(5, 3) = area: %.2f, perimeter: %.2f\n", area, perimeter)

	fmt.Println("\n=== 可变参数 / Variadic Functions ===")

	fmt.Println("sum(1, 2, 3) =", variableSum(1, 2, 3))
	fmt.Println("sum(1, 2, 3, 4, 5) =", variableSum(1, 2, 3, 4, 5))

	// 传递切片
	// Pass slice
	nums := []int{10, 20, 30}
	fmt.Println("sum(nums...) =", variableSum(nums...))

	fmt.Println("\n=== 匿名函数 / Anonymous Functions ===")

	// 匿名函数赋值给变量
	// Assign anonymous function to variable
	square := func(x int) int {
		return x * x
	}
	fmt.Println("square(5) =", square(5))

	// 立即执行
	// Immediately invoked
	result2 := func(x, y int) int {
		return x + y
	}(10, 20)
	fmt.Println("IIFE result =", result2)

	fmt.Println("\n=== 闭包 / Closures ===")

	// 闭包
	// Closures
	counter := makeCounter()
	fmt.Println("counter() =", counter())
	fmt.Println("counter() =", counter())
	fmt.Println("counter() =", counter())

	// 多个闭包实例
	// Multiple closure instances
	c1 := makeCounter()
	c2 := makeCounter()
	fmt.Println("\nc1():", c1(), c1())
	fmt.Println("c2():", c2(), c2(), c2())

	fmt.Println("\n=== 函数作为参数 / Functions as Arguments ===")

	// 函数作为参数
	// Function as argument
	numbers := []int{1, 2, 3, 4, 5}
	doubled := mapInts(numbers, func(x int) int {
		return x * 2
	})
	fmt.Println("doubled:", doubled)

	filtered := filterInts(numbers, func(x int) bool {
		return x > 2
	})
	fmt.Println("filtered (>2):", filtered)

	fmt.Println("\n=== defer 语句 / defer Statement ===")

	// defer 基础
	// defer basics
	deferExample()

	// defer 与返回值
	// defer with return value
	fmt.Println("\ndeferWithReturn() =", deferWithReturn())

	fmt.Println("\n=== 值传递与指针 / Value vs Pointer ===")

	// 值传递
	// Pass by value
	num := 10
	incrementValue(num)
	fmt.Println("After incrementValue:", num) // 仍然是 10

	// 指针传递
	// Pass by pointer
	incrementPointer(&num)
	fmt.Println("After incrementPointer:", num) // 11

	fmt.Println("\n=== 方法（预览）/ Methods (Preview) ===")

	p := Person{Name: "Alice", Age: 30}
	fmt.Println(p.Greet())
	fmt.Println("Before birthday:", p.Age)
	p.HaveBirthday()
	fmt.Println("After birthday:", p.Age)
}

// 简单函数
// Simple functions
func add(a, b int) int {
	return a + b
}

func multiply(a, b int) int {
	return a * b
}

func sayHello() {
	fmt.Println("Hello!")
}

func greet(name string) {
	fmt.Printf("Hello, %s!\n", name)
}

// 多返回值
// Multiple return values
func sumAndProduct(a, b int) (int, int) {
	return a + b, a * b
}

func divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, fmt.Errorf("division by zero")
	}
	return a / b, nil
}

// 命名返回值
// Named return values
func rectangle(width, height float64) (area, perimeter float64) {
	area = width * height
	perimeter = 2 * (width + height)
	return // 裸返回
}

// 可变参数
// Variadic function
func variableSum(nums ...int) int {
	total := 0
	for _, n := range nums {
		total += n
	}
	return total
}

// 闭包
// Closure
func makeCounter() func() int {
	count := 0
	return func() int {
		count++
		return count
	}
}

// 高阶函数
// Higher-order functions
func mapInts(nums []int, f func(int) int) []int {
	result := make([]int, len(nums))
	for i, n := range nums {
		result[i] = f(n)
	}
	return result
}

func filterInts(nums []int, predicate func(int) bool) []int {
	var result []int
	for _, n := range nums {
		if predicate(n) {
			result = append(result, n)
		}
	}
	return result
}

// defer 示例
// defer example
func deferExample() {
	fmt.Println("Start")
	defer fmt.Println("Deferred 1")
	defer fmt.Println("Deferred 2")
	fmt.Println("End")
	// 输出顺序: Start, End, Deferred 2, Deferred 1
}

func deferWithReturn() (result int) {
	defer func() {
		result++
	}()
	return 10 // 实际返回 11
}

// 值传递
// Pass by value
func incrementValue(x int) {
	x++
}

// 指针传递
// Pass by pointer
func incrementPointer(x *int) {
	*x++
}

// 方法
// Methods
type Person struct {
	Name string
	Age  int
}

func (p Person) Greet() string {
	return fmt.Sprintf("Hello, I'm %s!", p.Name)
}

func (p *Person) HaveBirthday() {
	p.Age++
}

// 模拟资源清理
// Simulated resource cleanup
func processFile() error {
	fmt.Println("Opening file...")
	defer func() {
		fmt.Println("Closing file...")
	}()

	// 模拟处理
	time.Sleep(10 * time.Millisecond)
	fmt.Println("Processing file...")

	return nil
}
