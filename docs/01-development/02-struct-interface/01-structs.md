# 结构体基础

本文档介绍 Go 语言中结构体的定义、初始化和基本使用。

## 目录

- [结构体定义](#结构体定义)
- [结构体初始化](#结构体初始化)
- [访问字段](#访问字段)
- [匿名结构体](#匿名结构体)
- [结构体比较](#结构体比较)
- [与 JavaScript/TypeScript 对比](#与-javascripttypescript-对比)

---

## 结构体定义

### 基本定义

```go
package main

import "fmt"

// 定义结构体
type Person struct {
    Name string
    Age  int
    City string
}

func main() {
    // 创建结构体实例
    p := Person{
        Name: "Alice",
        Age:  30,
        City: "Beijing",
    }
    fmt.Printf("%+v\n", p)
}
```

### 字段命名规则

```go
type User struct {
    ID        int       // 导出字段（公开）
    Name      string    // 导出字段
    email     string    // 未导出字段（私有）
    createdAt time.Time // 未导出字段
}
```

**规则**:
- 首字母大写：导出（公开），包外可访问
- 首字母小写：未导出（私有），仅包内可访问

### 空结构体

```go
// 空结构体不占用内存
type Empty struct{}

func main() {
    var e Empty
    fmt.Println("Size of Empty:", unsafe.Sizeof(e)) // 0

    // 常用于 map 实现 Set
    set := make(map[string]struct{})
    set["item"] = struct{}{}
}
```

---

## 结构体初始化

### 方式一：字段名初始化（推荐）

```go
p1 := Person{
    Name: "Bob",
    Age:  25,
    City: "Shanghai",
}
```

### 方式二：位置初始化

```go
// 必须按顺序提供所有字段
p2 := Person{"Charlie", 28, "Guangzhou"}
```

**注意**: 位置初始化在字段变化时容易出错，推荐使用字段名初始化。

### 方式三：零值初始化

```go
// 所有字段初始化为零值
var p3 Person
fmt.Println(p3.Name) // ""
fmt.Println(p3.Age)  // 0
```

### 方式四：new 函数

```go
// 返回指针
p4 := new(Person)
p4.Name = "David"
p4.Age = 35
```

### 方式五：取地址

```go
p5 := &Person{
    Name: "Eve",
    Age:  22,
}
```

### 工厂函数（推荐）

```go
// 构造函数模式
func NewPerson(name string, age int) *Person {
    return &Person{
        Name: name,
        Age:  age,
        City: "Unknown",
    }
}

// 带验证的构造函数
func NewPersonWithValidation(name string, age int) (*Person, error) {
    if name == "" {
        return nil, errors.New("name cannot be empty")
    }
    if age < 0 || age > 150 {
        return nil, errors.New("invalid age")
    }
    return &Person{Name: name, Age: age}, nil
}
```

---

## 访问字段

### 读取和修改

```go
p := Person{Name: "Frank", Age: 40}

// 读取
fmt.Println(p.Name)

// 修改
p.Age = 41
```

### 指针访问

```go
p := &Person{Name: "Grace", Age: 30}

// Go 自动解引用，两种写法等价
fmt.Println(p.Name)      // 推荐
fmt.Println((*p).Name)   // 显式解引用
```

### 结构体作为函数参数

```go
// 值传递 - 复制整个结构体
func printPerson(p Person) {
    fmt.Printf("%+v\n", p)
}

// 指针传递 - 只传递地址
func updateAge(p *Person, newAge int) {
    p.Age = newAge
}

func main() {
    person := Person{Name: "Henry", Age: 50}

    printPerson(person)        // 传递副本
    updateAge(&person, 51)     // 传递指针
    fmt.Println(person.Age)    // 51
}
```

---

## 匿名结构体

### 一次性使用的结构体

```go
// 不需要定义类型名
person := struct {
    Name string
    Age  int
}{
    Name: "Ivan",
    Age:  25,
}
fmt.Printf("%+v\n", person)
```

### 常见用途

```go
// 1. 配置结构
config := struct {
    Host string
    Port int
}{
    Host: "localhost",
    Port: 8080,
}

// 2. 测试数据
tests := []struct {
    input    string
    expected int
}{
    {"hello", 5},
    {"world", 5},
    {"go", 2},
}

for _, tt := range tests {
    if len(tt.input) != tt.expected {
        fmt.Printf("Failed: %s\n", tt.input)
    }
}

// 3. JSON 响应
response := struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
}{
    Code:    200,
    Message: "success",
}
```

---

## 结构体比较

### 可比较的结构体

```go
type Point struct {
    X, Y int
}

func main() {
    p1 := Point{1, 2}
    p2 := Point{1, 2}
    p3 := Point{2, 3}

    fmt.Println(p1 == p2) // true
    fmt.Println(p1 == p3) // false
}
```

### 不可比较的结构体

包含切片、映射或函数的结构体不可比较：

```go
type Data struct {
    Values []int // 切片不可比较
}

func main() {
    d1 := Data{Values: []int{1, 2, 3}}
    d2 := Data{Values: []int{1, 2, 3}}

    // fmt.Println(d1 == d2) // 编译错误

    // 使用 reflect.DeepEqual
    fmt.Println(reflect.DeepEqual(d1, d2)) // true
}
```

---

## 与 JavaScript/TypeScript 对比

### 类型定义对比

```typescript
// TypeScript
interface Person {
    name: string;
    age: number;
    city?: string;  // 可选字段
}

class User {
    constructor(
        public name: string,
        private age: number
    ) {}
}
```

```go
// Go
type Person struct {
    Name string
    Age  int
    City string
}

// Go 没有可选字段，使用指针或零值表示
type PersonWithOptional struct {
    Name string
    Age  int
    City *string // nil 表示未设置
}
```

### 初始化对比

```typescript
// TypeScript
const person: Person = {
    name: "Alice",
    age: 30
};

// 类实例化
const user = new User("Bob", 25);
```

```go
// Go - 结构体字面量
person := Person{
    Name: "Alice",
    Age:  30,
}

// Go - 工厂函数（没有 new 关键字用于自定义类型）
user := NewUser("Bob", 25)
```

### 访问控制对比

```typescript
// TypeScript
class User {
    public name: string;
    private age: number;
    protected email: string;
}
```

```go
// Go - 通过首字母大小写控制
type User struct {
    Name  string  // public (导出)
    age   int     // private (未导出)
    email string  // private (未导出)
}
```

### 主要差异

| 特性 | TypeScript | Go |
|-----|-----------|-----|
| 类型定义 | `class`, `interface` | `struct` |
| 可选字段 | `field?` | 指针或零值 |
| 访问控制 | 关键字 | 首字母大小写 |
| 构造函数 | `constructor` | 工厂函数 |
| 继承 | `extends` | 嵌入(后续章节) |
| 默认值 | 参数默认值 | 需要工厂函数 |

---

## 最佳实践

### 1. 使用工厂函数

```go
// 好的实践
func NewPerson(name string, age int) *Person {
    return &Person{
        Name: name,
        Age:  age,
    }
}
```

### 2. 字段名初始化

```go
// 好
p := Person{Name: "Alice", Age: 30}

// 避免
p := Person{"Alice", 30, "Beijing"}
```

### 3. 小结构体传值，大结构体传指针

```go
// 小结构体 - 传值
type Point struct {
    X, Y int
}

func distance(p1, p2 Point) float64 {
    // ...
}

// 大结构体 - 传指针
type LargeConfig struct {
    // 很多字段...
}

func processConfig(cfg *LargeConfig) {
    // ...
}
```

### 4. 合理组织字段顺序

```go
// 好 - 相关字段放在一起
type User struct {
    // 标识信息
    ID        int
    Username  string

    // 个人信息
    FirstName string
    LastName  string
    Age       int

    // 元数据
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

---

## 下一步

- [结构体标签](./02-struct-tags.md) - 学习使用 json, db 等标签
