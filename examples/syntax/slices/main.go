// Package main demonstrates arrays and slices in Go.
//
// To run this example:
//
//	go run main.go
//
// Expected output:
//
//	=== 数组 / Arrays ===
//	零值数组: [0 0 0 0 0]
//	...
package main

import "fmt"

func main() {
	fmt.Println("=== 数组 / Arrays ===")

	// 数组声明
	// Array declaration
	var arr1 [5]int
	fmt.Println("零值数组:", arr1)

	// 声明并初始化
	// Declare and initialize
	arr2 := [5]int{1, 2, 3, 4, 5}
	fmt.Println("初始化数组:", arr2)

	// 部分初始化
	// Partial initialization
	arr3 := [5]int{1, 2}
	fmt.Println("部分初始化:", arr3)

	// 让编译器计算长度
	// Let compiler calculate length
	arr4 := [...]int{1, 2, 3, 4, 5, 6}
	fmt.Println("自动长度:", arr4, "len:", len(arr4))

	// 指定索引初始化
	// Initialize by index
	arr5 := [5]int{1: 10, 3: 30}
	fmt.Println("按索引初始化:", arr5)

	// 数组是值类型
	// Arrays are value types
	fmt.Println("\n数组是值类型 / Arrays are Value Types:")
	original := [3]int{1, 2, 3}
	copied := original
	copied[0] = 100
	fmt.Println("原数组:", original)
	fmt.Println("复制后修改:", copied)

	fmt.Println("\n=== 切片 / Slices ===")

	// 创建切片
	// Create slices
	s1 := []int{1, 2, 3, 4, 5}
	fmt.Println("字面量创建:", s1)

	s2 := make([]int, 5)
	fmt.Printf("make([]int, 5): %v, len=%d, cap=%d\n", s2, len(s2), cap(s2))

	s3 := make([]int, 3, 10)
	fmt.Printf("make([]int, 3, 10): %v, len=%d, cap=%d\n", s3, len(s3), cap(s3))

	// 从数组创建切片
	// Create slice from array
	arr := [5]int{10, 20, 30, 40, 50}
	s4 := arr[1:4]
	fmt.Println("\n从数组创建切片 arr[1:4]:", s4)

	fmt.Println("\n=== 切片操作 / Slice Operations ===")

	// append 追加元素
	// append elements
	s := []int{1, 2, 3}
	fmt.Println("原切片:", s)

	s = append(s, 4)
	fmt.Println("append(s, 4):", s)

	s = append(s, 5, 6, 7)
	fmt.Println("append(s, 5, 6, 7):", s)

	// 追加另一个切片
	// Append another slice
	other := []int{8, 9, 10}
	s = append(s, other...)
	fmt.Println("append(s, other...):", s)

	// 切片表达式
	// Slice expressions
	data := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	fmt.Println("\ndata:", data)
	fmt.Println("data[2:5]:", data[2:5])
	fmt.Println("data[:3]:", data[:3])
	fmt.Println("data[7:]:", data[7:])
	fmt.Println("data[:]:", data[:])

	// copy 复制切片
	// copy slices
	src := []int{1, 2, 3}
	dst := make([]int, len(src))
	copied2 := copy(dst, src)
	fmt.Printf("\n复制了 %d 个元素: %v\n", copied2, dst)

	fmt.Println("\n=== 切片是引用类型 / Slices are Reference Types ===")

	slice1 := []int{1, 2, 3}
	slice2 := slice1
	slice2[0] = 100
	fmt.Println("slice1:", slice1)
	fmt.Println("slice2:", slice2)
	fmt.Println("修改 slice2 也会影响 slice1!")

	// 独立副本
	// Independent copy
	slice3 := make([]int, len(slice1))
	copy(slice3, slice1)
	slice3[0] = 999
	fmt.Println("\n创建独立副本:")
	fmt.Println("slice1:", slice1)
	fmt.Println("slice3:", slice3)

	fmt.Println("\n=== 删除元素 / Delete Elements ===")

	nums := []int{1, 2, 3, 4, 5}
	fmt.Println("原切片:", nums)

	// 删除索引 2 的元素
	// Delete element at index 2
	i := 2
	nums = append(nums[:i], nums[i+1:]...)
	fmt.Println("删除索引2后:", nums)

	// 删除第一个元素
	// Delete first element
	nums = nums[1:]
	fmt.Println("删除第一个:", nums)

	// 删除最后一个元素
	// Delete last element
	nums = nums[:len(nums)-1]
	fmt.Println("删除最后一个:", nums)

	fmt.Println("\n=== 容量与扩容 / Capacity and Growth ===")

	growth := make([]int, 0, 2)
	fmt.Printf("初始: len=%d, cap=%d\n", len(growth), cap(growth))

	for j := 0; j < 10; j++ {
		growth = append(growth, j)
		fmt.Printf("追加 %d: len=%d, cap=%d\n", j, len(growth), cap(growth))
	}

	fmt.Println("\n=== nil 切片 / Nil Slice ===")

	var nilSlice []int
	fmt.Println("nil 切片:", nilSlice)
	fmt.Println("是否为 nil:", nilSlice == nil)
	fmt.Println("长度:", len(nilSlice))
	fmt.Println("容量:", cap(nilSlice))

	// nil 切片可以安全地 append
	// nil slice can safely append
	nilSlice = append(nilSlice, 1, 2, 3)
	fmt.Println("append 后:", nilSlice)

	fmt.Println("\n=== 遍历切片 / Iterate Slice ===")

	items := []string{"apple", "banana", "orange"}

	// 使用 range
	// Using range
	fmt.Println("使用 range:")
	for i, v := range items {
		fmt.Printf("  [%d] %s\n", i, v)
	}

	// 只要索引
	// Index only
	fmt.Println("只要索引:")
	for i := range items {
		fmt.Printf("  index: %d\n", i)
	}

	// 只要值
	// Value only
	fmt.Println("只要值:")
	for _, v := range items {
		fmt.Printf("  value: %s\n", v)
	}
}
