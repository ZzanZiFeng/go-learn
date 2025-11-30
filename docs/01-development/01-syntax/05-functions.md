# 函数

本文档介绍 Go 语言的函数定义、参数、返回值和相关特性。

## 目录

- [函数定义](#函数定义)
- [参数](#参数)
- [返回值](#返回值)
- [匿名函数与闭包](#匿名函数与闭包)
- [defer 语句](#defer-语句)
- [与 JavaScript/TypeScript 对比](#与-javascripttypescript-对比)

---

## 函数定义

### 基本语法

```go
func functionName(param1 type1, param2 type2) returnType {
    // 函数体
    return value
}
```

```go
package main

import "fmt"

// 无参数无返回值
func sayHello() {
    fmt.Println("Hello!")
}

// 有参数有返回值
func add(a int, b int) int {
    return a + b
}

// 相同类型参数可以合并类型声明
func multiply(a, b int) int {
    return a * b
}

func main() {
    sayHello()
    fmt.Println("add(3, 5) =", add(3, 5))
    fmt.Println("multiply(3, 5) =", multiply(3, 5))
}
```

---

## 参数

### 值传递

Go 的函数参数都是值传递（传递副本）。

```go
package main

import "fmt"

func modifyValue(x int) {
    x = 100
    fmt.Println("Inside function:", x)
}

func main() {
    num := 10
    modifyValue(num)
    fmt.Println("After function:", num)  // 仍然是 10
}
```

### 指针参数

使用指针可以修改原始值。

```go
package main

import "fmt"

func modifyPointer(x *int) {
    *x = 100
}

func main() {
    num := 10
    modifyPointer(&num)
    fmt.Println("After function:", num)  // 100
}
```

### 切片和 Map 参数

切片和 Map 是引用类型，函数内修改会影响原数据。

```go
package main

import "fmt"

func modifySlice(s []int) {
    s[0] = 100
}

func modifyMap(m map[string]int) {
    m["key"] = 100
}

func main() {
    slice := []int{1, 2, 3}
    modifySlice(slice)
    fmt.Println("Slice:", slice)  // [100 2 3]

    m := map[string]int{"key": 1}
    modifyMap(m)
    fmt.Println("Map:", m)  // map[key:100]
}
```

### 可变参数

```go
package main

import "fmt"

// 可变参数
func sum(nums ...int) int {
    total := 0
    for _, n := range nums {
        total += n
    }
    return total
}

func main() {
    fmt.Println(sum(1, 2, 3))           // 6
    fmt.Println(sum(1, 2, 3, 4, 5))     // 15

    // 传递切片需要展开
    numbers := []int{10, 20, 30}
    fmt.Println(sum(numbers...))        // 60
}
```

---

## 返回值

### 多返回值

Go 函数可以返回多个值，这是处理错误的惯用方式。

```go
package main

import (
    "errors"
    "fmt"
)

// 多返回值
func divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, errors.New("division by zero")
    }
    return a / b, nil
}

func main() {
    // 接收多返回值
    result, err := divide(10, 2)
    if err != nil {
        fmt.Println("Error:", err)
    } else {
        fmt.Println("Result:", result)
    }

    // 忽略某个返回值
    result2, _ := divide(20, 4)
    fmt.Println("Result2:", result2)
}
```

### 命名返回值

```go
package main

import "fmt"

// 命名返回值
func rectangle(width, height float64) (area, perimeter float64) {
    area = width * height
    perimeter = 2 * (width + height)
    return  // 裸返回
}

// 也可以显式返回
func rectangleExplicit(width, height float64) (area, perimeter float64) {
    area = width * height
    perimeter = 2 * (width + height)
    return area, perimeter
}

func main() {
    a, p := rectangle(5, 3)
    fmt.Printf("Area: %.2f, Perimeter: %.2f\n", a, p)
}
```

**注意**: 命名返回值会初始化为零值，裸返回会返回当前值。在复杂函数中要谨慎使用裸返回。

---

## 匿名函数与闭包

### 匿名函数

```go
package main

import "fmt"

func main() {
    // 匿名函数赋值给变量
    add := func(a, b int) int {
        return a + b
    }
    fmt.Println("add(3, 5) =", add(3, 5))

    // 立即执行的匿名函数
    result := func(x int) int {
        return x * x
    }(5)
    fmt.Println("5² =", result)
}
```

### 闭包

闭包可以捕获外部变量。

```go
package main

import "fmt"

// 返回闭包
func counter() func() int {
    count := 0
    return func() int {
        count++
        return count
    }
}

func main() {
    // 每个闭包有自己的状态
    c1 := counter()
    c2 := counter()

    fmt.Println("c1:", c1(), c1(), c1())  // 1 2 3
    fmt.Println("c2:", c2(), c2())         // 1 2
}
```

### 闭包陷阱

```go
package main

import (
    "fmt"
    "time"
)

func main() {
    // 常见陷阱：循环变量捕获
    for i := 0; i < 3; i++ {
        go func() {
            fmt.Println(i)  // 可能全部打印 3
        }()
    }

    time.Sleep(time.Second)
    fmt.Println("---")

    // 正确方式 1: 传参
    for i := 0; i < 3; i++ {
        go func(n int) {
            fmt.Println(n)
        }(i)
    }

    time.Sleep(time.Second)
    fmt.Println("---")

    // 正确方式 2: 在循环内创建新变量
    for i := 0; i < 3; i++ {
        i := i  // 创建新变量
        go func() {
            fmt.Println(i)
        }()
    }

    time.Sleep(time.Second)
}
```

---

## defer 语句

`defer` 延迟执行函数调用，常用于资源清理。

### 基本用法

```go
package main

import "fmt"

func main() {
    fmt.Println("Start")
    defer fmt.Println("Deferred 1")
    defer fmt.Println("Deferred 2")
    defer fmt.Println("Deferred 3")
    fmt.Println("End")
}
```

输出:
```
Start
End
Deferred 3
Deferred 2
Deferred 1
```

**注意**: defer 以 LIFO（后进先出）顺序执行。

### 资源清理

```go
package main

import (
    "fmt"
    "os"
)

func readFile(filename string) error {
    file, err := os.Open(filename)
    if err != nil {
        return err
    }
    defer file.Close()  // 函数返回前关闭文件

    // 使用 file 进行操作...
    buf := make([]byte, 100)
    n, _ := file.Read(buf)
    fmt.Println("Read", n, "bytes")

    return nil
}

func main() {
    err := readFile("test.txt")
    if err != nil {
        fmt.Println("Error:", err)
    }
}
```

### defer 与返回值

```go
package main

import "fmt"

func deferReturn() (result int) {
    defer func() {
        result++  // 修改命名返回值
    }()
    return 10  // 实际返回 11
}

func main() {
    fmt.Println(deferReturn())  // 11
}
```

### defer 与 panic/recover

```go
package main

import "fmt"

func mayPanic() {
    defer func() {
        if r := recover(); r != nil {
            fmt.Println("Recovered from:", r)
        }
    }()
    panic("something went wrong")
}

func main() {
    mayPanic()
    fmt.Println("Program continues...")
}
```

---

## 与 JavaScript/TypeScript 对比

### 函数定义对比

```javascript
// JavaScript
function add(a, b) {
    return a + b;
}

// 箭头函数
const add = (a, b) => a + b;

// TypeScript
function add(a: number, b: number): number {
    return a + b;
}
```

```go
// Go
func add(a, b int) int {
    return a + b
}

// 没有箭头函数语法，但有匿名函数
add := func(a, b int) int {
    return a + b
}
```

### 多返回值对比

```javascript
// JavaScript - 使用对象或数组模拟
function divide(a, b) {
    if (b === 0) {
        return { result: null, error: "division by zero" };
    }
    return { result: a / b, error: null };
}

// 解构
const { result, error } = divide(10, 2);

// 或使用数组
function divide(a, b) {
    return b === 0 ? [null, "error"] : [a / b, null];
}
const [result, error] = divide(10, 2);
```

```go
// Go - 原生多返回值
func divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, errors.New("division by zero")
    }
    return a / b, nil
}

result, err := divide(10, 2)
```

### 可变参数对比

```javascript
// JavaScript
function sum(...nums) {
    return nums.reduce((a, b) => a + b, 0);
}

sum(1, 2, 3);
sum(...[1, 2, 3]);
```

```go
// Go
func sum(nums ...int) int {
    total := 0
    for _, n := range nums {
        total += n
    }
    return total
}

sum(1, 2, 3)
sum([]int{1, 2, 3}...)
```

### 闭包对比

```javascript
// JavaScript
function counter() {
    let count = 0;
    return function() {
        count++;
        return count;
    };
}

const c = counter();
console.log(c()); // 1
console.log(c()); // 2
```

```go
// Go
func counter() func() int {
    count := 0
    return func() int {
        count++
        return count
    }
}

c := counter()
fmt.Println(c()) // 1
fmt.Println(c()) // 2
```

### defer vs finally

```javascript
// JavaScript - try/finally
async function readFile(filename) {
    const file = await open(filename);
    try {
        // 使用文件
        return await file.read();
    } finally {
        await file.close();  // 总是执行
    }
}
```

```go
// Go - defer
func readFile(filename string) ([]byte, error) {
    file, err := os.Open(filename)
    if err != nil {
        return nil, err
    }
    defer file.Close()  // 函数返回前执行

    return io.ReadAll(file)
}
```

### 差异总结

| 特性 | JavaScript | Go |
|-----|-----------|-----|
| 函数声明 | `function`, 箭头函数 | `func` |
| 类型声明 | 可选 (TS必须) | 必须 |
| 多返回值 | 模拟 | 原生支持 |
| 可变参数 | `...args` | `...type` |
| 默认参数 | 支持 | 不支持 |
| 资源清理 | `try/finally` | `defer` |
| this 绑定 | 有 | 无 (方法用 receiver) |
| 函数重载 | 无 (可模拟) | 无 |

### Go 没有默认参数

```javascript
// JavaScript - 默认参数
function greet(name = "World") {
    console.log(`Hello, ${name}!`);
}
greet();        // Hello, World!
greet("Go");    // Hello, Go!
```

```go
// Go - 需要其他方式实现
func greet(name string) {
    if name == "" {
        name = "World"
    }
    fmt.Printf("Hello, %s!\n", name)
}

// 或使用可选配置模式
type GreetOptions struct {
    Name string
}

func greetWithOptions(opts ...GreetOptions) {
    name := "World"
    if len(opts) > 0 && opts[0].Name != "" {
        name = opts[0].Name
    }
    fmt.Printf("Hello, %s!\n", name)
}
```

---

## 最佳实践

1. **优先返回 error 而非 panic**: 除非是不可恢复的错误
2. **使用 defer 进行资源清理**: 确保资源释放
3. **命名返回值用于文档**: 但避免裸返回
4. **小函数优于大函数**: 单一职责原则
5. **避免过多参数**: 考虑使用结构体

```go
// 好的实践
func processUser(opts ProcessOptions) (*User, error) {
    if err := validate(opts); err != nil {
        return nil, err  // 返回错误而非 panic
    }

    file, err := os.Open(opts.ConfigPath)
    if err != nil {
        return nil, err
    }
    defer file.Close()  // defer 清理

    // 处理逻辑...
    return user, nil
}
```

---

## 下一步

- [指针](./06-pointers.md) - 学习指针的使用
