# 泛型基础

本文档介绍 Go 1.18+ 引入的泛型功能。

## 目录

- [泛型简介](#泛型简介)
- [泛型函数](#泛型函数)
- [泛型类型](#泛型类型)
- [类型约束](#类型约束)
- [内置约束](#内置约束)
- [泛型最佳实践](#泛型最佳实践)
- [与 JavaScript/TypeScript 对比](#与-javascripttypescript-对比)

---

## 泛型简介

### 什么是泛型？

泛型允许编写可以适用于多种类型的代码，同时保持类型安全。Go 1.18 开始支持泛型。

### 没有泛型时的问题

```go
// 没有泛型：需要为每种类型写重复代码
func MinInt(a, b int) int {
    if a < b {
        return a
    }
    return b
}

func MinFloat64(a, b float64) float64 {
    if a < b {
        return a
    }
    return b
}

func MinString(a, b string) string {
    if a < b {
        return a
    }
    return b
}

// 或者使用 interface{} 但失去类型安全
func MinAny(a, b interface{}) interface{} {
    // 需要运行时类型断言，可能 panic
    return nil
}
```

### 使用泛型的解决方案

```go
package main

import (
    "fmt"
    "cmp"
)

// 泛型函数：一个函数处理所有可比较类型
func Min[T cmp.Ordered](a, b T) T {
    if a < b {
        return a
    }
    return b
}

func main() {
    fmt.Println(Min(3, 5))           // 3 (int)
    fmt.Println(Min(3.14, 2.71))     // 2.71 (float64)
    fmt.Println(Min("apple", "banana")) // apple (string)
}
```

---

## 泛型函数

### 基本语法

```go
func FunctionName[T TypeConstraint](param T) T {
    // 函数体
}
```

- `T`: 类型参数名（约定使用大写字母）
- `TypeConstraint`: 类型约束（限制 T 可以是什么类型）

### 简单泛型函数

```go
package main

import "fmt"

// 泛型函数：交换两个值
func Swap[T any](a, b T) (T, T) {
    return b, a
}

func main() {
    x, y := Swap(1, 2)
    fmt.Println(x, y)  // 2 1

    s1, s2 := Swap("hello", "world")
    fmt.Println(s1, s2)  // world hello
}
```

### 多个类型参数

```go
func Map[T, U any](slice []T, f func(T) U) []U {
    result := make([]U, len(slice))
    for i, v := range slice {
        result[i] = f(v)
    }
    return result
}

func main() {
    nums := []int{1, 2, 3, 4, 5}

    // int -> string
    strings := Map(nums, func(n int) string {
        return fmt.Sprintf("num:%d", n)
    })
    fmt.Println(strings)  // [num:1 num:2 num:3 num:4 num:5]

    // int -> int
    doubled := Map(nums, func(n int) int {
        return n * 2
    })
    fmt.Println(doubled)  // [2 4 6 8 10]
}
```

### 类型推断

Go 可以自动推断类型参数：

```go
func First[T any](slice []T) T {
    return slice[0]
}

func main() {
    // 显式指定类型
    x := First[int]([]int{1, 2, 3})

    // 类型推断（推荐）
    y := First([]string{"a", "b", "c"})

    fmt.Println(x, y)  // 1 a
}
```

---

## 泛型类型

### 泛型结构体

```go
package main

import "fmt"

// 泛型栈
type Stack[T any] struct {
    items []T
}

func (s *Stack[T]) Push(item T) {
    s.items = append(s.items, item)
}

func (s *Stack[T]) Pop() (T, bool) {
    if len(s.items) == 0 {
        var zero T
        return zero, false
    }
    item := s.items[len(s.items)-1]
    s.items = s.items[:len(s.items)-1]
    return item, true
}

func (s *Stack[T]) Peek() (T, bool) {
    if len(s.items) == 0 {
        var zero T
        return zero, false
    }
    return s.items[len(s.items)-1], true
}

func (s *Stack[T]) Len() int {
    return len(s.items)
}

func main() {
    // 整数栈
    intStack := &Stack[int]{}
    intStack.Push(1)
    intStack.Push(2)
    intStack.Push(3)

    val, _ := intStack.Pop()
    fmt.Println(val)  // 3

    // 字符串栈
    strStack := &Stack[string]{}
    strStack.Push("hello")
    strStack.Push("world")

    str, _ := strStack.Pop()
    fmt.Println(str)  // world
}
```

### 泛型 Map

```go
type Pair[K comparable, V any] struct {
    Key   K
    Value V
}

type OrderedMap[K comparable, V any] struct {
    pairs []Pair[K, V]
}

func (m *OrderedMap[K, V]) Set(key K, value V) {
    for i, p := range m.pairs {
        if p.Key == key {
            m.pairs[i].Value = value
            return
        }
    }
    m.pairs = append(m.pairs, Pair[K, V]{Key: key, Value: value})
}

func (m *OrderedMap[K, V]) Get(key K) (V, bool) {
    for _, p := range m.pairs {
        if p.Key == key {
            return p.Value, true
        }
    }
    var zero V
    return zero, false
}

func main() {
    om := &OrderedMap[string, int]{}
    om.Set("one", 1)
    om.Set("two", 2)

    val, ok := om.Get("one")
    fmt.Println(val, ok)  // 1 true
}
```

---

## 类型约束

### 自定义约束

```go
// 定义约束接口
type Number interface {
    int | int8 | int16 | int32 | int64 |
    uint | uint8 | uint16 | uint32 | uint64 |
    float32 | float64
}

func Sum[T Number](nums []T) T {
    var sum T
    for _, n := range nums {
        sum += n
    }
    return sum
}

func main() {
    ints := []int{1, 2, 3, 4, 5}
    fmt.Println(Sum(ints))  // 15

    floats := []float64{1.1, 2.2, 3.3}
    fmt.Println(Sum(floats))  // 6.6
}
```

### 约束中的方法

```go
// 带方法的约束
type Stringer interface {
    String() string
}

func PrintAll[T Stringer](items []T) {
    for _, item := range items {
        fmt.Println(item.String())
    }
}

type Person struct {
    Name string
}

func (p Person) String() string {
    return "Person: " + p.Name
}

func main() {
    people := []Person{
        {Name: "Alice"},
        {Name: "Bob"},
    }
    PrintAll(people)
    // Person: Alice
    // Person: Bob
}
```

### 组合约束

```go
// 组合：类型集合 + 方法
type StringerNumber interface {
    Number
    String() string
}

// 使用 ~ 支持底层类型
type MyInt int

func (m MyInt) String() string {
    return fmt.Sprintf("MyInt(%d)", m)
}

// ~ 表示底层类型是 int 的所有类型
type IntLike interface {
    ~int | ~int8 | ~int16 | ~int32 | ~int64
}

func Double[T IntLike](n T) T {
    return n * 2
}

func main() {
    var m MyInt = 5
    fmt.Println(Double(m))  // 10
}
```

---

## 内置约束

### any 约束

```go
// any 是 interface{} 的别名
func Print[T any](v T) {
    fmt.Println(v)
}
```

### comparable 约束

```go
// comparable 表示支持 == 和 != 的类型
func Contains[T comparable](slice []T, target T) bool {
    for _, v := range slice {
        if v == target {
            return true
        }
    }
    return false
}

func main() {
    nums := []int{1, 2, 3, 4, 5}
    fmt.Println(Contains(nums, 3))   // true
    fmt.Println(Contains(nums, 10))  // false

    strs := []string{"a", "b", "c"}
    fmt.Println(Contains(strs, "b")) // true
}
```

### cmp.Ordered 约束

```go
import "cmp"

// cmp.Ordered 表示支持 < <= >= > 的类型
func Max[T cmp.Ordered](a, b T) T {
    if a > b {
        return a
    }
    return b
}

func Sort[T cmp.Ordered](slice []T) {
    // 简单冒泡排序
    for i := 0; i < len(slice)-1; i++ {
        for j := 0; j < len(slice)-1-i; j++ {
            if slice[j] > slice[j+1] {
                slice[j], slice[j+1] = slice[j+1], slice[j]
            }
        }
    }
}

func main() {
    fmt.Println(Max(3, 5))       // 5
    fmt.Println(Max("a", "z"))   // z

    nums := []int{5, 2, 8, 1, 9}
    Sort(nums)
    fmt.Println(nums)  // [1 2 5 8 9]
}
```

### constraints 包

```go
import "golang.org/x/exp/constraints"

// constraints.Integer - 所有整数类型
// constraints.Float - 所有浮点类型
// constraints.Signed - 有符号整数
// constraints.Unsigned - 无符号整数
// constraints.Complex - 复数类型
// constraints.Ordered - 可排序类型

func Abs[T constraints.Signed](n T) T {
    if n < 0 {
        return -n
    }
    return n
}
```

---

## 泛型最佳实践

### 1. 何时使用泛型

**适合使用泛型的场景**：
- 容器类型（栈、队列、列表）
- 通用算法（排序、搜索、过滤）
- 减少重复代码的函数

```go
// 好：通用的过滤函数
func Filter[T any](slice []T, predicate func(T) bool) []T {
    var result []T
    for _, v := range slice {
        if predicate(v) {
            result = append(result, v)
        }
    }
    return result
}
```

**不适合使用泛型的场景**：
- 只有一两种类型
- 不同类型需要不同处理逻辑

```go
// 不需要泛型：只处理字符串
func TrimAndLower(s string) string {
    return strings.ToLower(strings.TrimSpace(s))
}
```

### 2. 保持约束简单

```go
// 好：简单明确的约束
type Number interface {
    int | float64
}

// 避免：过于复杂的约束
type ComplexConstraint interface {
    ~int | ~int8 | ~int16 | ~int32 | ~int64 |
    ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
    ~float32 | ~float64 |
    String() string
    Compare(other any) int
}
```

### 3. 类型参数命名

```go
// 常用命名约定
T    - 通用类型 (Type)
K    - 键类型 (Key)
V    - 值类型 (Value)
E    - 元素类型 (Element)
S    - 切片类型 (Slice)

// 示例
func MapEntries[K comparable, V any](m map[K]V) []Pair[K, V] {
    // ...
}
```

### 4. 零值处理

```go
func SafeGet[T any](slice []T, index int) T {
    if index < 0 || index >= len(slice) {
        var zero T  // 返回零值
        return zero
    }
    return slice[index]
}

// 或返回指针表示可能为空
func SafeGetPtr[T any](slice []T, index int) *T {
    if index < 0 || index >= len(slice) {
        return nil
    }
    return &slice[index]
}
```

---

## 与 JavaScript/TypeScript 对比

### 基本泛型语法

```typescript
// TypeScript
function identity<T>(arg: T): T {
    return arg;
}

// 使用
const num = identity<number>(42);
const str = identity("hello");  // 类型推断
```

```go
// Go
func Identity[T any](arg T) T {
    return arg
}

// 使用
num := Identity[int](42)
str := Identity("hello")  // 类型推断
```

### 泛型约束

```typescript
// TypeScript
interface Lengthwise {
    length: number;
}

function logLength<T extends Lengthwise>(arg: T): void {
    console.log(arg.length);
}
```

```go
// Go
type Lengthwise interface {
    Len() int
}

func LogLength[T Lengthwise](arg T) {
    fmt.Println(arg.Len())
}
```

### 联合类型约束

```typescript
// TypeScript
function add<T extends string | number>(a: T, b: T): T {
    return (a as any) + (b as any);  // 需要类型断言
}
```

```go
// Go
type StringOrNumber interface {
    string | int | float64
}

func Add[T StringOrNumber](a, b T) T {
    return a + b  // 直接操作
}
```

### 泛型类/结构体

```typescript
// TypeScript
class Stack<T> {
    private items: T[] = [];

    push(item: T): void {
        this.items.push(item);
    }

    pop(): T | undefined {
        return this.items.pop();
    }
}
```

```go
// Go
type Stack[T any] struct {
    items []T
}

func (s *Stack[T]) Push(item T) {
    s.items = append(s.items, item)
}

func (s *Stack[T]) Pop() (T, bool) {
    if len(s.items) == 0 {
        var zero T
        return zero, false
    }
    item := s.items[len(s.items)-1]
    s.items = s.items[:len(s.items)-1]
    return item, true
}
```

### 主要差异

| 特性 | TypeScript | Go |
|-----|-----------|-----|
| 语法 | `<T>` | `[T constraint]` |
| 约束 | `extends` | 接口或类型集合 |
| 类型推断 | 强大 | 基本支持 |
| 默认类型 | 支持 `<T = string>` | 不支持 |
| 条件类型 | 支持 | 不支持 |
| 方差注解 | `in/out` | 不支持 |

---

## 常用泛型工具函数

```go
package main

import (
    "cmp"
    "fmt"
)

// 返回切片中的第一个元素
func First[T any](slice []T) (T, bool) {
    if len(slice) == 0 {
        var zero T
        return zero, false
    }
    return slice[0], true
}

// 返回切片中的最后一个元素
func Last[T any](slice []T) (T, bool) {
    if len(slice) == 0 {
        var zero T
        return zero, false
    }
    return slice[len(slice)-1], true
}

// 过滤切片
func Filter[T any](slice []T, predicate func(T) bool) []T {
    result := make([]T, 0)
    for _, v := range slice {
        if predicate(v) {
            result = append(result, v)
        }
    }
    return result
}

// 映射切片
func Map[T, U any](slice []T, mapper func(T) U) []U {
    result := make([]U, len(slice))
    for i, v := range slice {
        result[i] = mapper(v)
    }
    return result
}

// 归约切片
func Reduce[T, U any](slice []T, initial U, reducer func(U, T) U) U {
    result := initial
    for _, v := range slice {
        result = reducer(result, v)
    }
    return result
}

// 查找元素
func Find[T any](slice []T, predicate func(T) bool) (T, bool) {
    for _, v := range slice {
        if predicate(v) {
            return v, true
        }
    }
    var zero T
    return zero, false
}

// 查找最大值
func Max[T cmp.Ordered](slice []T) (T, bool) {
    if len(slice) == 0 {
        var zero T
        return zero, false
    }
    max := slice[0]
    for _, v := range slice[1:] {
        if v > max {
            max = v
        }
    }
    return max, true
}

// 查找最小值
func Min[T cmp.Ordered](slice []T) (T, bool) {
    if len(slice) == 0 {
        var zero T
        return zero, false
    }
    min := slice[0]
    for _, v := range slice[1:] {
        if v < min {
            min = v
        }
    }
    return min, true
}

func main() {
    nums := []int{1, 2, 3, 4, 5}

    // Filter
    evens := Filter(nums, func(n int) bool { return n%2 == 0 })
    fmt.Println("Evens:", evens)  // [2 4]

    // Map
    doubled := Map(nums, func(n int) int { return n * 2 })
    fmt.Println("Doubled:", doubled)  // [2 4 6 8 10]

    // Reduce
    sum := Reduce(nums, 0, func(acc, n int) int { return acc + n })
    fmt.Println("Sum:", sum)  // 15

    // Max/Min
    max, _ := Max(nums)
    min, _ := Min(nums)
    fmt.Println("Max:", max, "Min:", min)  // Max: 5 Min: 1
}
```

---

## 下一步

恭喜你完成了结构体与接口章节！接下来：

- [并发编程](../03-concurrency/) - 学习 Go 的 goroutine 和 channel
