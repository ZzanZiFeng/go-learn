# 基本类型

本文档介绍 Go 语言的基本数据类型。

## 目录

- [数值类型](#数值类型)
- [字符串](#字符串)
- [布尔类型](#布尔类型)
- [类型转换](#类型转换)
- [与 JavaScript/TypeScript 对比](#与-javascripttypescript-对比)

---

## 数值类型

### 整数类型

| 类型 | 大小 | 范围 |
|-----|------|------|
| `int8` | 8 bit | -128 ~ 127 |
| `int16` | 16 bit | -32768 ~ 32767 |
| `int32` | 32 bit | -2^31 ~ 2^31-1 |
| `int64` | 64 bit | -2^63 ~ 2^63-1 |
| `int` | 平台相关 | 32位系统为int32，64位为int64 |

无符号整数：

| 类型 | 大小 | 范围 |
|-----|------|------|
| `uint8` (byte) | 8 bit | 0 ~ 255 |
| `uint16` | 16 bit | 0 ~ 65535 |
| `uint32` | 32 bit | 0 ~ 2^32-1 |
| `uint64` | 64 bit | 0 ~ 2^64-1 |
| `uint` | 平台相关 | 32位系统为uint32，64位为uint64 |

```go
package main

import "fmt"

func main() {
    var a int = 42
    var b int64 = 9223372036854775807
    var c uint8 = 255
    var d byte = 65  // byte 是 uint8 的别名

    fmt.Printf("int: %d\n", a)
    fmt.Printf("int64: %d\n", b)
    fmt.Printf("uint8: %d\n", c)
    fmt.Printf("byte as char: %c\n", d)  // 输出 'A'
}
```

输出:
```
int: 42
int64: 9223372036854775807
uint8: 255
byte as char: A
```

### 浮点数类型

| 类型 | 大小 | 精度 |
|-----|------|------|
| `float32` | 32 bit | ~6-7 位有效数字 |
| `float64` | 64 bit | ~15-16 位有效数字 |

```go
package main

import (
    "fmt"
    "math"
)

func main() {
    var f32 float32 = 3.14159265358979
    var f64 float64 = 3.14159265358979

    fmt.Printf("float32: %.10f\n", f32)  // 精度丢失
    fmt.Printf("float64: %.10f\n", f64)  // 更精确

    // 特殊值
    fmt.Println("Max float64:", math.MaxFloat64)
    fmt.Println("Smallest positive float64:", math.SmallestNonzeroFloat64)
}
```

输出:
```
float32: 3.1415927410
float64: 3.1415926536
Max float64: 1.7976931348623157e+308
Smallest positive float64: 5e-324
```

### 复数类型

Go 原生支持复数（JavaScript 不支持）：

```go
package main

import (
    "fmt"
    "math/cmplx"
)

func main() {
    var c64 complex64 = 1 + 2i
    var c128 complex128 = 3 + 4i

    fmt.Println("complex64:", c64)
    fmt.Println("complex128:", c128)

    // 复数运算
    fmt.Println("Sum:", c64+complex64(c128))
    fmt.Println("Abs:", cmplx.Abs(c128))  // 模
}
```

---

## 字符串

### 字符串基础

Go 的字符串是不可变的 UTF-8 编码的字节序列。

```go
package main

import "fmt"

func main() {
    // 字符串声明
    s1 := "Hello, World!"
    s2 := "你好，世界！"

    fmt.Println(s1)
    fmt.Println(s2)

    // 字符串长度
    fmt.Println("s1 bytes:", len(s1))  // 字节数
    fmt.Println("s2 bytes:", len(s2))  // UTF-8 字节数（中文3字节/字符）

    // 使用 rune 计算字符数
    fmt.Println("s2 runes:", len([]rune(s2)))
}
```

输出:
```
Hello, World!
你好，世界！
s1 bytes: 13
s2 bytes: 18
s2 runes: 6
```

### 字符串操作

```go
package main

import (
    "fmt"
    "strings"
)

func main() {
    s := "Hello, World!"

    // 索引（返回字节，不是字符）
    fmt.Println("First byte:", s[0])         // 72 ('H')
    fmt.Printf("As char: %c\n", s[0])        // H

    // 切片
    fmt.Println("Slice [0:5]:", s[0:5])      // Hello
    fmt.Println("Slice [7:]:", s[7:])        // World!

    // 字符串拼接
    greeting := "Hello" + ", " + "Gopher"
    fmt.Println(greeting)

    // strings 包常用函数
    fmt.Println("Contains:", strings.Contains(s, "World"))
    fmt.Println("HasPrefix:", strings.HasPrefix(s, "Hello"))
    fmt.Println("ToUpper:", strings.ToUpper(s))
    fmt.Println("Split:", strings.Split(s, ", "))
    fmt.Println("Replace:", strings.Replace(s, "World", "Go", 1))
}
```

### 原始字符串

```go
package main

import "fmt"

func main() {
    // 普通字符串 - 支持转义字符
    s1 := "Line1\nLine2\tTabbed"

    // 原始字符串 - 不处理转义
    s2 := `Line1\nLine2\tTabbed`

    // 多行字符串
    s3 := `
    This is a
    multi-line
    string
    `

    fmt.Println("Normal string:")
    fmt.Println(s1)
    fmt.Println("\nRaw string:")
    fmt.Println(s2)
    fmt.Println("\nMulti-line:")
    fmt.Println(s3)
}
```

### rune 类型

`rune` 是 `int32` 的别名，表示一个 Unicode 码点。

```go
package main

import "fmt"

func main() {
    s := "Hello, 世界"

    // 遍历字节
    fmt.Println("Bytes:")
    for i := 0; i < len(s); i++ {
        fmt.Printf("%d: %c\n", i, s[i])
    }

    // 遍历 rune（字符）
    fmt.Println("\nRunes:")
    for i, r := range s {
        fmt.Printf("%d: %c (U+%04X)\n", i, r, r)
    }
}
```

---

## 布尔类型

```go
package main

import "fmt"

func main() {
    var b1 bool = true
    var b2 bool = false
    b3 := true

    fmt.Println("b1:", b1)
    fmt.Println("b2:", b2)
    fmt.Println("b3:", b3)

    // 比较运算
    x, y := 10, 20
    fmt.Println("x == y:", x == y)
    fmt.Println("x < y:", x < y)
    fmt.Println("x != y:", x != y)

    // 逻辑运算
    fmt.Println("true && false:", true && false)
    fmt.Println("true || false:", true || false)
    fmt.Println("!true:", !true)
}
```

**注意**: Go 的布尔值不能与整数互换，不能用 0/1 代替 false/true。

```go
// 错误示例
// if 1 { }        // 编译错误: non-bool 1 used as if condition
// var b bool = 1  // 编译错误: cannot use 1 as bool

// 正确方式
if true { }
var b bool = true
```

---

## 类型转换

Go 不支持隐式类型转换，必须显式转换。

```go
package main

import (
    "fmt"
    "strconv"
)

func main() {
    // 数值类型转换
    var i int = 42
    var f float64 = float64(i)
    var u uint = uint(f)

    fmt.Printf("int: %d, float64: %f, uint: %d\n", i, f, u)

    // 字符串与数值转换
    // int -> string
    s1 := strconv.Itoa(42)
    fmt.Println("int to string:", s1)

    // string -> int
    n, err := strconv.Atoi("42")
    if err != nil {
        fmt.Println("Error:", err)
    } else {
        fmt.Println("string to int:", n)
    }

    // float64 -> string
    s2 := strconv.FormatFloat(3.14159, 'f', 2, 64)
    fmt.Println("float to string:", s2)

    // string -> float64
    f2, _ := strconv.ParseFloat("3.14159", 64)
    fmt.Println("string to float:", f2)

    // 使用 fmt.Sprintf
    s3 := fmt.Sprintf("%d", 42)
    fmt.Println("Sprintf:", s3)
}
```

### 类型别名

```go
package main

import "fmt"

// 类型别名
type MyInt int
type UserID int64

func main() {
    var a int = 10
    var b MyInt = 20

    // 需要显式转换
    c := MyInt(a) + b
    fmt.Println("c:", c)

    // 实际应用
    var userID UserID = 12345
    fmt.Printf("User ID: %d\n", userID)
}
```

---

## 与 JavaScript/TypeScript 对比

### 数值类型对比

```javascript
// JavaScript - 只有 number 和 BigInt
let num = 42;              // number (64位浮点数)
let big = 9007199254740993n; // BigInt

// 没有整数类型，大整数精度问题
console.log(9007199254740993);  // 9007199254740992 (精度丢失!)
```

```typescript
// TypeScript - 类型注解但底层仍是 JS
let num: number = 42;
let big: bigint = 9007199254740993n;
```

```go
// Go - 明确的整数类型
var i int = 42
var i64 int64 = 9007199254740993  // 无精度问题
var f float64 = 3.14
```

### 字符串对比

```javascript
// JavaScript
let s = "Hello";
s[0] = 'h';     // 静默失败，字符串不可变
console.log(s); // "Hello"

// 模板字符串
let name = "World";
let greeting = `Hello, ${name}!`;

// 字符串长度
"Hello".length;      // 5
"你好".length;        // 2 (字符数)
```

```go
// Go
s := "Hello"
// s[0] = 'h'  // 编译错误: 字符串不可变

// 格式化字符串（没有模板字符串）
name := "World"
greeting := fmt.Sprintf("Hello, %s!", name)

// 字符串长度
len("Hello")           // 5 (字节数)
len("你好")            // 6 (字节数，不是字符数！)
len([]rune("你好"))    // 2 (字符数)
```

### 布尔类型对比

```javascript
// JavaScript - 真值/假值概念
if (1) console.log("truthy");
if ("hello") console.log("truthy");
if (0) console.log("won't print");
if ("") console.log("won't print");
if (null) console.log("won't print");
if (undefined) console.log("won't print");

// 隐式转换
let b = !!1;  // true
```

```go
// Go - 只有 true/false
// if 1 { }           // 编译错误
// if "hello" { }     // 编译错误
if true { }           // OK
if len("hello") > 0 { }  // OK - 显式比较
```

### 类型转换对比

```javascript
// JavaScript - 隐式转换（coercion）
console.log("5" + 3);     // "53" (字符串拼接)
console.log("5" - 3);     // 2 (数字运算)
console.log("5" * "3");   // 15

// 显式转换
Number("42");             // 42
String(42);               // "42"
parseInt("42px");         // 42
```

```go
// Go - 无隐式转换
// fmt.Println("5" + 3)  // 编译错误

// 必须显式转换
n, _ := strconv.Atoi("42")  // 42
s := strconv.Itoa(42)       // "42"
```

### 差异总结

| 特性 | JavaScript | Go |
|-----|-----------|-----|
| 数值类型 | number (64位浮点) | int, int64, float64 等 |
| 大整数 | BigInt | int64, big.Int |
| 字符串 | 16位 UTF-16 | UTF-8 字节序列 |
| 字符串长度 | 字符数 | 字节数 |
| 字符串遍历 | for...of (字符) | range (rune) |
| 布尔转换 | 真值/假值 | 只有 true/false |
| 类型转换 | 隐式 + 显式 | 只有显式 |

---

## 最佳实践

1. **默认使用 `int` 和 `float64`**: 除非有特殊需求
2. **处理大整数用 `int64`**: 避免溢出问题
3. **字符串处理注意 UTF-8**: 使用 `range` 或 `[]rune` 遍历字符
4. **显式类型转换**: 不要期望隐式转换

```go
// 好的实践
count := 100              // 使用 int
price := 19.99            // 使用 float64
id := int64(userInput)    // 明确转换

// 处理多语言字符串
for _, r := range text {
    // r 是 rune，正确处理 UTF-8
}
```

---

## 下一步

- [复合类型](./03-composite-types.md) - 学习数组、切片和映射
