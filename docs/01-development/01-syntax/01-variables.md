# 变量与常量

本文档介绍 Go 语言中变量和常量的声明与使用方式。

## 目录

- [变量声明](#变量声明)
- [短变量声明](#短变量声明)
- [常量](#常量)
- [零值](#零值)
- [类型推断](#类型推断)
- [与 JavaScript/TypeScript 对比](#与-javascripttypescript-对比)

---

## 变量声明

### 使用 var 关键字

```go
// 声明单个变量
var name string
var age int

// 声明并初始化
var name string = "Gopher"
var age int = 25

// 类型推断（省略类型）
var name = "Gopher"
var age = 25

// 声明多个同类型变量
var x, y, z int

// 声明多个不同类型变量
var (
    name   string = "Gopher"
    age    int    = 25
    active bool   = true
)
```

**运行示例:**

```go
package main

import "fmt"

func main() {
    var name string = "Gopher"
    var age int = 25

    fmt.Println("Name:", name)
    fmt.Println("Age:", age)
}
```

输出:
```
Name: Gopher
Age: 25
```

---

## 短变量声明

### 使用 := 操作符

`:=` 是 Go 特有的简短声明语法，只能在函数内部使用。

```go
// 短变量声明（自动推断类型）
name := "Gopher"
age := 25
price := 19.99
isActive := true

// 同时声明多个变量
x, y := 10, 20
name, age := "Alice", 30
```

**运行示例:**

```go
package main

import "fmt"

func main() {
    // 短变量声明
    name := "Gopher"
    age := 25

    // 同时声明多个
    x, y := 10, 20

    fmt.Println(name, age)
    fmt.Printf("x = %d, y = %d\n", x, y)
}
```

输出:
```
Gopher 25
x = 10, y = 20
```

### var vs := 使用场景

| 场景 | 推荐方式 | 原因 |
|-----|---------|------|
| 函数内部声明 | `:=` | 简洁 |
| 包级别声明 | `var` | 不能使用 := |
| 显式指定类型 | `var` | 更清晰 |
| 声明但不初始化 | `var` | := 必须初始化 |

```go
// 包级别变量（只能用 var）
var GlobalConfig = "production"

func main() {
    // 函数内部（推荐 :=）
    name := "Gopher"

    // 需要显式类型时用 var
    var count int64 = 100

    // 声明但不初始化时用 var
    var result string
}
```

---

## 常量

### 使用 const 关键字

常量在编译时确定值，不能被修改。

```go
// 单个常量
const Pi = 3.14159
const Greeting = "Hello"

// 显式类型
const MaxSize int = 100

// 常量组
const (
    StatusOK       = 200
    StatusNotFound = 404
    StatusError    = 500
)
```

### iota 枚举器

`iota` 是 Go 的常量计数器，每遇到一个新的 `const` 块重置为 0。

```go
const (
    Sunday    = iota  // 0
    Monday           // 1
    Tuesday          // 2
    Wednesday        // 3
    Thursday         // 4
    Friday           // 5
    Saturday         // 6
)

// 跳过值
const (
    _  = iota  // 0 (跳过)
    KB = 1 << (10 * iota)  // 1 << 10 = 1024
    MB                      // 1 << 20
    GB                      // 1 << 30
    TB                      // 1 << 40
)
```

**运行示例:**

```go
package main

import "fmt"

const (
    StatusOK    = 200
    StatusError = 500
)

const (
    _  = iota
    KB = 1 << (10 * iota)
    MB
    GB
)

func main() {
    fmt.Println("Status OK:", StatusOK)
    fmt.Println("KB:", KB)
    fmt.Println("MB:", MB)
    fmt.Println("GB:", GB)
}
```

输出:
```
Status OK: 200
KB: 1024
MB: 1048576
GB: 1073741824
```

---

## 零值

Go 的变量如果声明后未初始化，会自动赋予"零值"。

| 类型 | 零值 |
|-----|------|
| int, int8, int16... | `0` |
| float32, float64 | `0.0` |
| bool | `false` |
| string | `""` (空字符串) |
| pointer, slice, map, channel, func, interface | `nil` |

```go
package main

import "fmt"

func main() {
    var i int
    var f float64
    var b bool
    var s string

    fmt.Printf("int: %d\n", i)       // 0
    fmt.Printf("float64: %f\n", f)   // 0.000000
    fmt.Printf("bool: %t\n", b)      // false
    fmt.Printf("string: %q\n", s)    // ""
}
```

输出:
```
int: 0
float64: 0.000000
bool: false
string: ""
```

---

## 类型推断

Go 编译器可以根据值自动推断变量类型。

```go
// 整数默认推断为 int
x := 42        // int

// 浮点数默认推断为 float64
y := 3.14      // float64

// 字符串
s := "hello"   // string

// 布尔值
b := true      // bool

// 如果需要其他类型，必须显式声明
var i32 int32 = 42
var f32 float32 = 3.14
```

### 检查变量类型

```go
package main

import "fmt"

func main() {
    x := 42
    y := 3.14
    s := "hello"

    fmt.Printf("x: type=%T, value=%v\n", x, x)
    fmt.Printf("y: type=%T, value=%v\n", y, y)
    fmt.Printf("s: type=%T, value=%v\n", s, s)
}
```

输出:
```
x: type=int, value=42
y: type=float64, value=3.14
s: type=string, value=hello
```

---

## 与 JavaScript/TypeScript 对比

### 变量声明对比

```javascript
// JavaScript
let name = "Gopher";      // 可变变量
const age = 25;           // 常量
var legacy = "old";       // 旧式声明（避免使用）

// 重新赋值
name = "Alice";           // OK
// age = 30;              // Error: Assignment to constant variable
```

```typescript
// TypeScript
let name: string = "Gopher";
const age: number = 25;

// 类型推断
let inferred = "Hello";   // 推断为 string
```

```go
// Go
var name string = "Gopher"
const age = 25

// 短变量声明
name := "Gopher"

// 重新赋值
name = "Alice"            // OK
// age = 30               // Error: cannot assign to age
```

### 主要差异

| 特性 | JavaScript/TypeScript | Go |
|-----|----------------------|-----|
| 块作用域 | `let`, `const` | `var`, `:=` |
| 函数作用域 | `var` (旧式) | - |
| 常量 | `const` (引用不可变) | `const` (值不可变) |
| 类型推断 | 有 (TS) | 有 |
| 零值 | `undefined` | 类型相关的零值 |
| 声明后必须使用 | 否 | 是 (编译错误) |

### const 行为差异

```javascript
// JavaScript const - 引用不可变
const arr = [1, 2, 3];
arr.push(4);              // OK! 数组内容可以修改
// arr = [5, 6];          // Error: 引用不能改

const obj = { name: "Go" };
obj.name = "Gopher";      // OK! 属性可以修改
```

```go
// Go const - 值必须在编译时确定，不能是引用类型
const x = 10              // OK
const s = "hello"         // OK

// 以下都是错误的:
// const arr = []int{1,2,3}  // Error: 不能用于 slice
// const m = map[string]int{} // Error: 不能用于 map
```

### 未使用变量

```javascript
// JavaScript - 未使用的变量只是警告（可配置）
let unused = 10;
console.log("done");
```

```go
// Go - 未使用的变量是编译错误
func main() {
    unused := 10
    fmt.Println("done")
    // 编译错误: unused declared but not used
}
```

解决方案：

```go
// 使用空白标识符
unused := getValue()
_ = unused  // 明确忽略

// 或者直接忽略返回值
_ = getValue()
```

---

## 最佳实践

1. **优先使用 `:=`**: 在函数内部，短变量声明更简洁
2. **需要零值时用 `var`**: 当你需要变量的零值时
3. **组织相关变量**: 使用 `var()` 或 `const()` 块组织相关声明
4. **命名规范**:
   - 驼峰命名: `userName`, `maxSize`
   - 首字母大写表示导出: `MaxSize`, `UserName`
   - 短名称用于小作用域: `i`, `n`, `err`

```go
// 好的实践
var (
    defaultTimeout = 30 * time.Second
    maxRetries     = 3
)

func process() error {
    name := getUserName()
    if name == "" {
        return errors.New("name is empty")
    }
    // ...
}
```

---

## 下一步

- [基本类型](./02-basic-types.md) - 学习 Go 的基本数据类型
