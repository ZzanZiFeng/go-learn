# 指针

本文档介绍 Go 语言中指针的概念和使用方法。

## 目录

- [指针基础](#指针基础)
- [指针操作](#指针操作)
- [指针与函数](#指针与函数)
- [指针与结构体](#指针与结构体)
- [与 JavaScript/TypeScript 对比](#与-javascripttypescript-对比)

---

## 指针基础

### 什么是指针？

指针是存储另一个变量内存地址的变量。

```
变量 x  → 值: 10
          地址: 0xc000018030

指针 p  → 值: 0xc000018030 (x 的地址)
          地址: 0xc000018038
```

### 声明和使用

```go
package main

import "fmt"

func main() {
    x := 10

    // & 取地址运算符
    p := &x
    fmt.Printf("x 的值: %d\n", x)
    fmt.Printf("x 的地址: %p\n", &x)
    fmt.Printf("p 的值 (x的地址): %p\n", p)
    fmt.Printf("p 指向的值: %d\n", *p)  // * 解引用

    // 通过指针修改值
    *p = 20
    fmt.Printf("修改后 x 的值: %d\n", x)  // 20
}
```

输出:
```
x 的值: 10
x 的地址: 0xc000018030
p 的值 (x的地址): 0xc000018030
p 指向的值: 10
修改后 x 的值: 20
```

### 指针的零值

指针的零值是 `nil`。

```go
package main

import "fmt"

func main() {
    var p *int
    fmt.Println("p is nil:", p == nil)  // true

    // 解引用 nil 指针会 panic
    // fmt.Println(*p)  // panic: runtime error

    // 使用前检查
    if p != nil {
        fmt.Println(*p)
    }
}
```

### 使用 new 创建指针

```go
package main

import "fmt"

func main() {
    // new 分配内存并返回指针
    p := new(int)  // 分配 int，初始化为零值
    fmt.Println("*p =", *p)  // 0

    *p = 100
    fmt.Println("*p =", *p)  // 100

    // 等价于
    var x int
    p2 := &x
    fmt.Println("*p2 =", *p2)  // 0
}
```

---

## 指针操作

### 基本运算符

| 运算符 | 名称 | 说明 |
|--------|------|------|
| `&` | 取地址 | 获取变量的内存地址 |
| `*` | 解引用 | 获取指针指向的值 |

```go
package main

import "fmt"

func main() {
    x := 42

    // & 取地址
    p := &x
    fmt.Printf("p 类型: %T\n", p)  // *int

    // * 解引用
    v := *p
    fmt.Printf("v 类型: %T, 值: %d\n", v, v)  // int, 42

    // * 用于类型声明表示指针类型
    var p2 *int = &x
    fmt.Println("*p2 =", *p2)
}
```

### 指针与切片/Map

切片和 Map 本身就是引用类型，通常不需要使用指针。

```go
package main

import "fmt"

func main() {
    // 切片本身就是引用
    s := []int{1, 2, 3}
    modifySlice(s)
    fmt.Println("s:", s)  // [100 2 3]

    // Map 本身就是引用
    m := map[string]int{"a": 1}
    modifyMap(m)
    fmt.Println("m:", m)  // map[a:100]
}

func modifySlice(s []int) {
    s[0] = 100
}

func modifyMap(m map[string]int) {
    m["a"] = 100
}
```

---

## 指针与函数

### 值传递 vs 指针传递

```go
package main

import "fmt"

// 值传递 - 无法修改原值
func incrementValue(x int) {
    x++
}

// 指针传递 - 可以修改原值
func incrementPointer(x *int) {
    (*x)++
}

func main() {
    a := 10
    incrementValue(a)
    fmt.Println("After incrementValue:", a)  // 10 (未改变)

    incrementPointer(&a)
    fmt.Println("After incrementPointer:", a)  // 11 (改变了)
}
```

### 何时使用指针参数

```go
package main

import "fmt"

// 1. 需要修改参数值时
func swap(a, b *int) {
    *a, *b = *b, *a
}

// 2. 避免大结构体的复制
type LargeStruct struct {
    Data [1000]int
}

func processLarge(s *LargeStruct) {
    // 只传递指针，不复制整个结构体
    s.Data[0] = 100
}

// 3. 表示可选参数 (nil = 未提供)
func greet(name *string) {
    if name == nil {
        fmt.Println("Hello, Guest!")
    } else {
        fmt.Printf("Hello, %s!\n", *name)
    }
}

func main() {
    x, y := 10, 20
    swap(&x, &y)
    fmt.Printf("x=%d, y=%d\n", x, y)  // x=20, y=10

    large := &LargeStruct{}
    processLarge(large)
    fmt.Println("First element:", large.Data[0])

    greet(nil)
    name := "Gopher"
    greet(&name)
}
```

### 返回指针

```go
package main

import "fmt"

// 返回局部变量的指针是安全的
// Go 会自动将其分配到堆上（逃逸分析）
func createInt() *int {
    x := 42
    return &x  // 安全！
}

func createPerson() *Person {
    p := Person{Name: "Gopher", Age: 25}
    return &p  // 安全！
}

type Person struct {
    Name string
    Age  int
}

func main() {
    p := createInt()
    fmt.Println("*p =", *p)

    person := createPerson()
    fmt.Printf("Person: %+v\n", *person)
}
```

---

## 指针与结构体

### 结构体指针

```go
package main

import "fmt"

type Person struct {
    Name string
    Age  int
}

func main() {
    // 创建结构体指针
    p := &Person{Name: "Alice", Age: 30}

    // 访问字段 - Go 自动解引用
    fmt.Println(p.Name)   // 等价于 (*p).Name
    fmt.Println(p.Age)

    // 修改字段
    p.Age = 31
    fmt.Println("New age:", p.Age)

    // 使用 new
    p2 := new(Person)
    p2.Name = "Bob"
    p2.Age = 25
    fmt.Printf("p2: %+v\n", *p2)
}
```

### 方法接收者：值 vs 指针

```go
package main

import "fmt"

type Counter struct {
    Value int
}

// 值接收者 - 不能修改原对象
func (c Counter) IncrementValue() {
    c.Value++  // 只修改副本
}

// 指针接收者 - 可以修改原对象
func (c *Counter) IncrementPointer() {
    c.Value++  // 修改原对象
}

func (c *Counter) GetValue() int {
    return c.Value
}

func main() {
    c := Counter{Value: 0}

    c.IncrementValue()
    fmt.Println("After IncrementValue:", c.Value)  // 0

    c.IncrementPointer()
    fmt.Println("After IncrementPointer:", c.Value)  // 1

    // Go 自动转换：值调用指针方法
    c.IncrementPointer()  // 等价于 (&c).IncrementPointer()
    fmt.Println("After second increment:", c.Value)  // 2
}
```

### 指针接收者的选择

何时使用指针接收者：
1. 需要修改接收者
2. 接收者是大结构体
3. 一致性：如果某些方法用指针，所有方法都用

```go
type Buffer struct {
    data []byte
}

// 所有方法都使用指针接收者（一致性）
func (b *Buffer) Write(p []byte) {
    b.data = append(b.data, p...)
}

func (b *Buffer) String() string {
    return string(b.data)
}

func (b *Buffer) Len() int {
    return len(b.data)
}
```

---

## 与 JavaScript/TypeScript 对比

### 引用概念对比

```javascript
// JavaScript - 没有显式指针
// 原始类型是值传递
let x = 10;
let y = x;
y = 20;
console.log(x);  // 10 (未改变)

// 对象是引用传递
let obj1 = { value: 10 };
let obj2 = obj1;
obj2.value = 20;
console.log(obj1.value);  // 20 (改变了!)

// 无法获取变量的"地址"
```

```go
// Go - 显式控制值/引用
x := 10
y := x
y = 20
fmt.Println(x)  // 10 (未改变)

// 使用指针实现引用
a := 10
b := &a
*b = 20
fmt.Println(a)  // 20 (改变了)

// 可以获取地址
fmt.Printf("%p\n", &a)
```

### 函数参数对比

```javascript
// JavaScript
function modifyObject(obj) {
    obj.value = 100;  // 会修改原对象
}

function modifyPrimitive(x) {
    x = 100;  // 不会修改原值
}

let obj = { value: 1 };
modifyObject(obj);
console.log(obj.value);  // 100

let num = 1;
modifyPrimitive(num);
console.log(num);  // 1
```

```go
// Go
func modifyStruct(s MyStruct) {
    s.Value = 100  // 不会修改原结构体（值传递）
}

func modifyStructPointer(s *MyStruct) {
    s.Value = 100  // 会修改原结构体
}

s := MyStruct{Value: 1}
modifyStruct(s)
fmt.Println(s.Value)  // 1

modifyStructPointer(&s)
fmt.Println(s.Value)  // 100
```

### 差异总结

| 特性 | JavaScript | Go |
|-----|-----------|-----|
| 显式指针 | 无 | 有 (`*`, `&`) |
| 原始类型传递 | 值 | 值 |
| 对象传递 | 引用 | 值（需要指针才是引用） |
| 空引用 | `null`, `undefined` | `nil` |
| 获取地址 | 不可能 | `&variable` |
| 解引用 | 自动 | 显式 `*pointer` |
| 算术运算 | N/A | 不支持 |

### Go 中没有指针运算

```c
// C 语言 - 可以指针运算
int arr[3] = {1, 2, 3};
int *p = arr;
p++;  // 移动到下一个元素
```

```go
// Go - 没有指针运算
arr := [3]int{1, 2, 3}
p := &arr[0]
// p++  // 编译错误！

// 使用索引访问
for i := range arr {
    fmt.Println(arr[i])
}
```

---

## 最佳实践

### 何时使用指针

| 场景 | 使用指针 | 原因 |
|-----|---------|------|
| 需要修改参数 | 是 | 值传递无法修改 |
| 大结构体参数 | 是 | 避免复制开销 |
| 可选参数 | 是 | nil 表示未提供 |
| 接口一致性 | 是 | 如果有些方法需要指针 |
| 小结构体参数 | 否 | 复制成本低 |
| 只读访问 | 否 | 不需要修改 |

### 避免的陷阱

```go
// 陷阱 1: nil 指针解引用
var p *int
// *p = 10  // panic!

// 正确做法
if p != nil {
    *p = 10
}

// 陷阱 2: 返回局部变量指针后立即丢弃
// (实际上这在 Go 中是安全的，但要理解原因)
p := createInt()  // 安全，Go 会处理

// 陷阱 3: 不必要的指针
type SmallStruct struct {
    X, Y int
}
// 小结构体直接传值更高效
func distance(a, b SmallStruct) float64 {
    // ...
}
```

### 代码示例

```go
// 好的实践
type Config struct {
    Host     string
    Port     int
    Timeout  time.Duration
    MaxConns int
}

// 大结构体使用指针
func NewServer(cfg *Config) *Server {
    return &Server{
        config: cfg,
    }
}

// 使用指针接收者
func (s *Server) Start() error {
    // ...
}

// 小类型直接传值
func Add(a, b int) int {
    return a + b
}
```

---

## 下一步

- [错误处理](./07-error-handling.md) - 学习 Go 的错误处理方式
