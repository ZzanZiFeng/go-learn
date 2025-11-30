# 控制流

本文档介绍 Go 语言的控制流语句：if、for、switch。

## 目录

- [if 语句](#if-语句)
- [for 循环](#for-循环)
- [switch 语句](#switch-语句)
- [与 JavaScript/TypeScript 对比](#与-javascripttypescript-对比)

---

## if 语句

### 基本用法

```go
package main

import "fmt"

func main() {
    x := 10

    // 基本 if
    if x > 5 {
        fmt.Println("x is greater than 5")
    }

    // if-else
    if x > 15 {
        fmt.Println("x is greater than 15")
    } else {
        fmt.Println("x is not greater than 15")
    }

    // if-else if-else
    if x < 5 {
        fmt.Println("x is less than 5")
    } else if x < 10 {
        fmt.Println("x is less than 10")
    } else {
        fmt.Println("x is 10 or more")
    }
}
```

### 带初始化语句的 if

Go 允许在条件前执行一个简单语句，变量作用域限于 if 块内。

```go
package main

import (
    "fmt"
    "os"
)

func main() {
    // 初始化语句; 条件
    if num := 10; num > 5 {
        fmt.Println("num is greater than 5:", num)
    }
    // fmt.Println(num)  // 编译错误: num 作用域在 if 块内

    // 常见用法: 错误处理
    if err := doSomething(); err != nil {
        fmt.Println("Error:", err)
        os.Exit(1)
    }
}

func doSomething() error {
    return nil
}
```

### if 与错误处理

```go
package main

import (
    "fmt"
    "os"
)

func main() {
    // 典型的错误处理模式
    file, err := os.Open("test.txt")
    if err != nil {
        fmt.Println("Error opening file:", err)
        return
    }
    defer file.Close()

    // 使用 file...
    fmt.Println("File opened successfully")
}
```

---

## for 循环

Go 只有 `for` 一种循环结构，但它可以表达多种循环方式。

### 基本 for 循环

```go
package main

import "fmt"

func main() {
    // 传统 for 循环 (类似 C/Java)
    for i := 0; i < 5; i++ {
        fmt.Print(i, " ")
    }
    fmt.Println()

    // 可以省略初始化和后置语句 (类似 while)
    j := 0
    for j < 5 {
        fmt.Print(j, " ")
        j++
    }
    fmt.Println()

    // 无限循环
    // for {
    //     fmt.Println("forever")
    // }
}
```

### range 循环

```go
package main

import "fmt"

func main() {
    // 遍历切片
    nums := []int{10, 20, 30}
    for i, v := range nums {
        fmt.Printf("index=%d, value=%d\n", i, v)
    }

    // 只要值
    for _, v := range nums {
        fmt.Print(v, " ")
    }
    fmt.Println()

    // 只要索引
    for i := range nums {
        fmt.Print(i, " ")
    }
    fmt.Println()

    // 遍历 map
    m := map[string]int{"a": 1, "b": 2, "c": 3}
    for k, v := range m {
        fmt.Printf("%s: %d\n", k, v)
    }

    // 遍历字符串 (按 rune)
    s := "Hello, 世界"
    for i, r := range s {
        fmt.Printf("%d: %c\n", i, r)
    }

    // 遍历 channel
    ch := make(chan int, 3)
    ch <- 1
    ch <- 2
    ch <- 3
    close(ch)
    for v := range ch {
        fmt.Print(v, " ")
    }
    fmt.Println()
}
```

### break 和 continue

```go
package main

import "fmt"

func main() {
    // break 终止循环
    for i := 0; i < 10; i++ {
        if i == 5 {
            break
        }
        fmt.Print(i, " ")
    }
    fmt.Println()  // 0 1 2 3 4

    // continue 跳过当前迭代
    for i := 0; i < 10; i++ {
        if i%2 == 0 {
            continue
        }
        fmt.Print(i, " ")
    }
    fmt.Println()  // 1 3 5 7 9
}
```

### 带标签的 break/continue

```go
package main

import "fmt"

func main() {
    // 标签用于跳出嵌套循环
outer:
    for i := 0; i < 3; i++ {
        for j := 0; j < 3; j++ {
            if i == 1 && j == 1 {
                break outer  // 跳出外层循环
            }
            fmt.Printf("(%d, %d) ", i, j)
        }
    }
    fmt.Println()
}
```

---

## switch 语句

### 基本 switch

```go
package main

import "fmt"

func main() {
    day := 3

    switch day {
    case 1:
        fmt.Println("Monday")
    case 2:
        fmt.Println("Tuesday")
    case 3:
        fmt.Println("Wednesday")
    case 4:
        fmt.Println("Thursday")
    case 5:
        fmt.Println("Friday")
    case 6, 7:  // 多个值
        fmt.Println("Weekend")
    default:
        fmt.Println("Invalid day")
    }
}
```

**重要**: Go 的 switch 默认不会 fall through，不需要 break。

### 无条件 switch

```go
package main

import (
    "fmt"
    "time"
)

func main() {
    hour := time.Now().Hour()

    // 无条件 switch（类似 if-else if-else）
    switch {
    case hour < 12:
        fmt.Println("Good morning!")
    case hour < 17:
        fmt.Println("Good afternoon!")
    case hour < 21:
        fmt.Println("Good evening!")
    default:
        fmt.Println("Good night!")
    }
}
```

### 带初始化的 switch

```go
package main

import "fmt"

func main() {
    switch num := 15; {
    case num < 0:
        fmt.Println("Negative")
    case num == 0:
        fmt.Println("Zero")
    case num > 0:
        fmt.Println("Positive")
    }
}
```

### fallthrough

```go
package main

import "fmt"

func main() {
    n := 2

    switch n {
    case 1:
        fmt.Println("One")
        fallthrough
    case 2:
        fmt.Println("Two")
        fallthrough
    case 3:
        fmt.Println("Three")
    case 4:
        fmt.Println("Four")
    }
    // 输出:
    // Two
    // Three
}
```

### 类型 switch

```go
package main

import "fmt"

func checkType(i interface{}) {
    switch v := i.(type) {
    case int:
        fmt.Printf("Integer: %d\n", v)
    case string:
        fmt.Printf("String: %s\n", v)
    case bool:
        fmt.Printf("Boolean: %t\n", v)
    case []int:
        fmt.Printf("Int slice: %v\n", v)
    default:
        fmt.Printf("Unknown type: %T\n", v)
    }
}

func main() {
    checkType(42)
    checkType("hello")
    checkType(true)
    checkType([]int{1, 2, 3})
    checkType(3.14)
}
```

---

## 与 JavaScript/TypeScript 对比

### if 语句对比

```javascript
// JavaScript
if (condition) {
    // ...
} else if (anotherCondition) {
    // ...
} else {
    // ...
}

// 条件可以是任意值（truthy/falsy）
if (1) { }           // true
if ("hello") { }     // true
if (null) { }        // false
```

```go
// Go
if condition {
    // ...
} else if anotherCondition {
    // ...
} else {
    // ...
}

// 条件必须是布尔值
// if 1 { }           // 编译错误
// if "hello" { }     // 编译错误
if true { }          // OK
if len("hello") > 0 { }  // OK
```

### 循环对比

```javascript
// JavaScript 有多种循环
// for 循环
for (let i = 0; i < 5; i++) { }

// while 循环
while (condition) { }

// do-while 循环
do { } while (condition);

// for...of (遍历值)
for (const item of array) { }

// for...in (遍历键)
for (const key in object) { }

// forEach 方法
array.forEach((item, index) => { });
```

```go
// Go 只有 for，但可以表达所有形式
// 传统 for
for i := 0; i < 5; i++ { }

// while 风格
for condition { }

// 无限循环
for { }

// range (类似 for...of)
for i, v := range slice { }

// 没有 forEach，但可以用 range
for _, item := range array {
    // ...
}
```

### switch 对比

```javascript
// JavaScript switch
switch (value) {
    case 1:
        console.log("one");
        break;  // 必须 break，否则 fall through
    case 2:
        console.log("two");
        break;
    default:
        console.log("other");
}

// 没有 break 会 fall through
switch (2) {
    case 1:
    case 2:
        console.log("one or two");  // 会执行
        // fall through!
    case 3:
        console.log("three");       // 也会执行！
}
```

```go
// Go switch - 默认不 fall through
switch value {
case 1:
    fmt.Println("one")
    // 自动 break
case 2:
    fmt.Println("two")
default:
    fmt.Println("other")
}

// 多个值
switch value {
case 1, 2:
    fmt.Println("one or two")
case 3:
    fmt.Println("three")
}

// 显式 fallthrough
switch 2 {
case 1:
    fallthrough
case 2:
    fmt.Println("one or two")
    fallthrough
case 3:
    fmt.Println("three")  // 需要 fallthrough 才执行
}
```

### 差异总结

| 特性 | JavaScript | Go |
|-----|-----------|-----|
| 循环种类 | for, while, do-while, for...of, for...in | 只有 for |
| if 条件 | 任意值 (truthy/falsy) | 必须布尔值 |
| switch fall through | 默认 fall through | 默认不 fall through |
| switch 表达式 | 必须有 | 可选 |
| 带初始化的 if | 不支持 | 支持 |
| 标签跳转 | 支持但不推荐 | 支持 |
| 三元运算符 | `a ? b : c` | 不支持 |

### 没有三元运算符

```javascript
// JavaScript
const result = condition ? valueIfTrue : valueIfFalse;
```

```go
// Go - 必须使用 if
var result int
if condition {
    result = valueIfTrue
} else {
    result = valueIfFalse
}

// 或者使用辅助函数
func ternary(condition bool, a, b int) int {
    if condition {
        return a
    }
    return b
}
result := ternary(x > 0, 1, -1)
```

---

## 最佳实践

1. **使用带初始化的 if**: 限制变量作用域
2. **优先使用 range**: 遍历集合时更简洁
3. **避免深层嵌套**: 早返回减少嵌套
4. **switch 优于多重 if-else**: 更清晰

```go
// 好的实践 - 早返回
func process(data string) error {
    if data == "" {
        return errors.New("empty data")
    }
    if len(data) > 100 {
        return errors.New("data too long")
    }
    // 主要逻辑
    return nil
}

// 好的实践 - switch 代替多重 if
func getStatusText(code int) string {
    switch code {
    case 200:
        return "OK"
    case 404:
        return "Not Found"
    case 500:
        return "Internal Server Error"
    default:
        return "Unknown"
    }
}
```

---

## 下一步

- [函数](./05-functions.md) - 学习函数定义和调用
