// Package main demonstrates pointers in Go.
//
// To run this example:
//
//	go run main.go
//
// Expected output:
//
//	=== 指针基础 / Pointer Basics ===
//	x 的值: 42
//	...
package main

import "fmt"

func main() {
	fmt.Println("=== 指针基础 / Pointer Basics ===")

	x := 42

	// & 取地址
	// & gets address
	p := &x

	fmt.Printf("x 的值: %d\n", x)
	fmt.Printf("x 的地址: %p\n", &x)
	fmt.Printf("p 的值 (x的地址): %p\n", p)
	fmt.Printf("p 的类型: %T\n", p)
	fmt.Printf("*p (解引用): %d\n", *p)

	// 通过指针修改值
	// Modify value through pointer
	*p = 100
	fmt.Printf("修改后 x 的值: %d\n", x)

	fmt.Println("\n=== 指针的零值 / Zero Value of Pointer ===")

	var nilPtr *int
	fmt.Printf("nil 指针: %v\n", nilPtr)
	fmt.Printf("是否为 nil: %t\n", nilPtr == nil)

	// 安全检查
	// Safe check
	if nilPtr != nil {
		fmt.Println(*nilPtr)
	} else {
		fmt.Println("指针为 nil，不能解引用")
	}

	fmt.Println("\n=== new 函数 / new Function ===")

	// 使用 new 分配内存
	// Allocate memory with new
	ptr := new(int)
	fmt.Printf("new(int) 值: %d\n", *ptr)
	fmt.Printf("new(int) 地址: %p\n", ptr)

	*ptr = 50
	fmt.Printf("赋值后: %d\n", *ptr)

	fmt.Println("\n=== 值传递 vs 指针传递 / Value vs Pointer ===")

	num := 10
	fmt.Println("原始值:", num)

	// 值传递 - 不能修改原值
	// Pass by value - cannot modify original
	incrementByValue(num)
	fmt.Println("incrementByValue 后:", num)

	// 指针传递 - 可以修改原值
	// Pass by pointer - can modify original
	incrementByPointer(&num)
	fmt.Println("incrementByPointer 后:", num)

	fmt.Println("\n=== 切片和 Map 是引用类型 / Slices and Maps are Reference Types ===")

	slice := []int{1, 2, 3}
	fmt.Println("原始切片:", slice)
	modifySlice(slice)
	fmt.Println("修改后切片:", slice)

	m := map[string]int{"a": 1}
	fmt.Println("原始 map:", m)
	modifyMap(m)
	fmt.Println("修改后 map:", m)

	fmt.Println("\n=== 结构体指针 / Struct Pointers ===")

	// 创建结构体
	// Create struct
	person := Person{Name: "Alice", Age: 30}
	fmt.Printf("原始 person: %+v\n", person)

	// 结构体指针
	// Struct pointer
	personPtr := &person

	// Go 自动解引用
	// Go auto-dereferences
	fmt.Println("personPtr.Name:", personPtr.Name)           // 等价于 (*personPtr).Name
	fmt.Println("(*personPtr).Name:", (*personPtr).Name)

	// 通过指针修改
	// Modify through pointer
	personPtr.Age = 31
	fmt.Printf("修改后 person: %+v\n", person)

	// 使用 new 创建结构体指针
	// Create struct pointer with new
	person2 := new(Person)
	person2.Name = "Bob"
	person2.Age = 25
	fmt.Printf("new(Person): %+v\n", *person2)

	fmt.Println("\n=== 方法接收者 / Method Receivers ===")

	counter := Counter{Value: 0}

	// 值接收者 - 不修改原对象
	// Value receiver - doesn't modify original
	counter.IncrementValue()
	fmt.Println("IncrementValue 后:", counter.Value)

	// 指针接收者 - 修改原对象
	// Pointer receiver - modifies original
	counter.IncrementPointer()
	fmt.Println("IncrementPointer 后:", counter.Value)

	// Go 自动转换
	// Go auto-converts
	counter.IncrementPointer() // 等价于 (&counter).IncrementPointer()
	fmt.Println("再次 IncrementPointer 后:", counter.Value)

	fmt.Println("\n=== 返回局部变量指针 / Return Local Variable Pointer ===")

	// 安全：Go 会将变量分配到堆上
	// Safe: Go allocates on heap
	ptr2 := createInt(100)
	fmt.Println("createInt 返回:", *ptr2)

	person3 := createPerson("Charlie", 35)
	fmt.Printf("createPerson 返回: %+v\n", *person3)

	fmt.Println("\n=== 指针数组 vs 数组指针 / Pointer Array vs Array Pointer ===")

	// 指针数组 - 数组的元素是指针
	// Array of pointers - elements are pointers
	a, b, c := 1, 2, 3
	ptrArray := [3]*int{&a, &b, &c}
	fmt.Println("指针数组 [3]*int:")
	for i, p := range ptrArray {
		fmt.Printf("  [%d] 地址=%p, 值=%d\n", i, p, *p)
	}

	// 数组指针 - 指向数组的指针
	// Pointer to array - pointer to an array
	arr := [3]int{10, 20, 30}
	arrayPtr := &arr
	fmt.Println("数组指针 *[3]int:")
	fmt.Printf("  arrayPtr[0] = %d\n", arrayPtr[0])

	fmt.Println("\n=== 双重指针 / Double Pointer ===")

	val := 10
	p1 := &val   // *int
	p2 := &p1    // **int

	fmt.Printf("val = %d\n", val)
	fmt.Printf("*p1 = %d\n", *p1)
	fmt.Printf("**p2 = %d\n", **p2)

	**p2 = 20
	fmt.Printf("修改后 val = %d\n", val)
}

// 值传递
func incrementByValue(x int) {
	x++
}

// 指针传递
func incrementByPointer(x *int) {
	*x++
}

// 修改切片（引用类型）
func modifySlice(s []int) {
	s[0] = 100
}

// 修改 map（引用类型）
func modifyMap(m map[string]int) {
	m["b"] = 2
}

// Person 结构体
type Person struct {
	Name string
	Age  int
}

// Counter 结构体
type Counter struct {
	Value int
}

// 值接收者
func (c Counter) IncrementValue() {
	c.Value++
}

// 指针接收者
func (c *Counter) IncrementPointer() {
	c.Value++
}

// 返回局部变量指针
func createInt(val int) *int {
	x := val
	return &x // 安全！Go 会进行逃逸分析
}

func createPerson(name string, age int) *Person {
	p := Person{Name: name, Age: age}
	return &p // 安全！
}
