// Package main demonstrates maps in Go.
//
// To run this example:
//
//	go run main.go
//
// Expected output:
//
//	=== Map 创建 / Map Creation ===
//	字面量创建: map[apple:5 banana:3 orange:7]
//	...
package main

import "fmt"

func main() {
	fmt.Println("=== Map 创建 / Map Creation ===")

	// 使用 make 创建
	// Create with make
	m1 := make(map[string]int)
	m1["one"] = 1
	m1["two"] = 2
	fmt.Println("make 创建:", m1)

	// 字面量创建
	// Literal creation
	m2 := map[string]int{
		"apple":  5,
		"banana": 3,
		"orange": 7,
	}
	fmt.Println("字面量创建:", m2)

	// 空 map
	// Empty map
	m3 := map[string]int{}
	fmt.Println("空 map:", m3, "是否为空:", len(m3) == 0)

	// nil map (只读)
	// nil map (read-only)
	var m4 map[string]int
	fmt.Println("nil map:", m4, "是否为 nil:", m4 == nil)
	// m4["key"] = 1  // panic: assignment to entry in nil map

	fmt.Println("\n=== Map 操作 / Map Operations ===")

	fruits := map[string]int{
		"apple":  5,
		"banana": 3,
		"orange": 7,
	}

	// 读取值
	// Read value
	fmt.Println("apple 数量:", fruits["apple"])

	// 读取不存在的键返回零值
	// Reading non-existent key returns zero value
	fmt.Println("grape 数量:", fruits["grape"]) // 0

	// 检查键是否存在
	// Check if key exists
	value, exists := fruits["banana"]
	if exists {
		fmt.Println("banana 存在，数量:", value)
	}

	value2, exists2 := fruits["mango"]
	if !exists2 {
		fmt.Println("mango 不存在，value:", value2) // 零值
	}

	// 添加/修改
	// Add/Update
	fruits["grape"] = 10
	fruits["apple"] = 15
	fmt.Println("添加/修改后:", fruits)

	// 删除
	// Delete
	delete(fruits, "banana")
	fmt.Println("删除 banana 后:", fruits)

	// 获取长度
	// Get length
	fmt.Println("Map 长度:", len(fruits))

	fmt.Println("\n=== 遍历 Map / Iterate Map ===")

	colors := map[string]string{
		"red":   "#FF0000",
		"green": "#00FF00",
		"blue":  "#0000FF",
	}

	// 遍历键值对
	// Iterate key-value pairs
	fmt.Println("遍历键值对:")
	for key, value := range colors {
		fmt.Printf("  %s: %s\n", key, value)
	}

	// 只遍历键
	// Iterate keys only
	fmt.Println("只遍历键:")
	for key := range colors {
		fmt.Printf("  key: %s\n", key)
	}

	// 注意：遍历顺序是随机的
	// Note: iteration order is random
	fmt.Println("\n遍历顺序是随机的，多次运行可能不同")

	fmt.Println("\n=== Map 是引用类型 / Map is Reference Type ===")

	original := map[string]int{"a": 1, "b": 2}
	reference := original
	reference["c"] = 3

	fmt.Println("original:", original)
	fmt.Println("reference:", reference)
	fmt.Println("修改 reference 也会影响 original!")

	fmt.Println("\n=== 嵌套 Map / Nested Map ===")

	// 嵌套 map
	// Nested map
	users := map[string]map[string]string{
		"user1": {
			"name":  "Alice",
			"email": "alice@example.com",
		},
		"user2": {
			"name":  "Bob",
			"email": "bob@example.com",
		},
	}

	fmt.Println("嵌套 map:")
	for userId, info := range users {
		fmt.Printf("  %s: name=%s, email=%s\n",
			userId, info["name"], info["email"])
	}

	// 安全访问嵌套 map
	// Safe access to nested map
	if user, ok := users["user1"]; ok {
		if name, ok := user["name"]; ok {
			fmt.Println("user1 的名字:", name)
		}
	}

	fmt.Println("\n=== Map 作为集合 / Map as Set ===")

	// 使用 map[T]bool 或 map[T]struct{} 作为集合
	// Use map[T]bool or map[T]struct{} as set

	// 方法 1: map[T]bool
	set1 := map[string]bool{
		"apple":  true,
		"banana": true,
	}

	// 检查元素是否存在
	if set1["apple"] {
		fmt.Println("apple 在 set1 中")
	}

	// 添加元素
	set1["orange"] = true

	// 删除元素
	delete(set1, "banana")

	fmt.Println("set1:", set1)

	// 方法 2: map[T]struct{} (更省内存)
	set2 := make(map[string]struct{})
	set2["apple"] = struct{}{}
	set2["banana"] = struct{}{}

	if _, exists := set2["apple"]; exists {
		fmt.Println("apple 在 set2 中")
	}

	fmt.Println("\n=== 常见用例 / Common Use Cases ===")

	// 统计词频
	// Word frequency
	text := []string{"go", "is", "fun", "go", "is", "powerful", "go"}
	wordCount := make(map[string]int)
	for _, word := range text {
		wordCount[word]++
	}
	fmt.Println("词频统计:", wordCount)

	// 分组
	// Grouping
	numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	groups := map[string][]int{
		"odd":  {},
		"even": {},
	}
	for _, n := range numbers {
		if n%2 == 0 {
			groups["even"] = append(groups["even"], n)
		} else {
			groups["odd"] = append(groups["odd"], n)
		}
	}
	fmt.Println("奇偶分组:", groups)

	// 缓存/记忆化
	// Cache/Memoization
	cache := make(map[int]int)
	fib := func(n int) int {
		if v, ok := cache[n]; ok {
			return v
		}
		// 计算并缓存
		result := n // 简化示例
		cache[n] = result
		return result
	}
	_ = fib(10)
	fmt.Println("缓存:", cache)
}
