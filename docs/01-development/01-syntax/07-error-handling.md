# 错误处理

本文档介绍 Go 语言的错误处理机制：error、panic 和 recover。

## 目录

- [error 接口](#error-接口)
- [创建错误](#创建错误)
- [错误处理模式](#错误处理模式)
- [panic 和 recover](#panic-和-recover)
- [与 JavaScript/TypeScript 对比](#与-javascripttypescript-对比)

---

## error 接口

### error 基础

Go 使用返回值而非异常来处理错误。`error` 是一个内置接口：

```go
type error interface {
    Error() string
}
```

```go
package main

import (
    "errors"
    "fmt"
    "os"
)

func main() {
    // 打开文件可能返回错误
    file, err := os.Open("nonexistent.txt")
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    defer file.Close()

    // 使用 file...
}
```

### 检查错误

```go
package main

import (
    "fmt"
    "strconv"
)

func main() {
    // 标准错误检查模式
    num, err := strconv.Atoi("42")
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    fmt.Println("Number:", num)

    // 转换失败的情况
    num2, err := strconv.Atoi("not a number")
    if err != nil {
        fmt.Println("Error:", err)
        // Error: strconv.Atoi: parsing "not a number": invalid syntax
    }

    // 多个返回值中只关心错误
    _, err = strconv.Atoi("invalid")
    if err != nil {
        fmt.Println("Conversion failed")
    }
}
```

---

## 创建错误

### 使用 errors.New

```go
package main

import (
    "errors"
    "fmt"
)

func divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, errors.New("division by zero")
    }
    return a / b, nil
}

func main() {
    result, err := divide(10, 0)
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    fmt.Println("Result:", result)
}
```

### 使用 fmt.Errorf

```go
package main

import (
    "fmt"
)

func getUser(id int) (string, error) {
    if id <= 0 {
        return "", fmt.Errorf("invalid user id: %d", id)
    }
    if id > 1000 {
        return "", fmt.Errorf("user id %d not found", id)
    }
    return "User" + fmt.Sprint(id), nil
}

func main() {
    user, err := getUser(-1)
    if err != nil {
        fmt.Println("Error:", err)  // Error: invalid user id: -1
    }

    user, err = getUser(50)
    if err != nil {
        fmt.Println("Error:", err)
    } else {
        fmt.Println("User:", user)
    }
}
```

### 自定义错误类型

```go
package main

import (
    "fmt"
)

// 自定义错误类型
type ValidationError struct {
    Field   string
    Message string
}

// 实现 error 接口
func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
}

func validateUser(name string, age int) error {
    if name == "" {
        return &ValidationError{
            Field:   "name",
            Message: "cannot be empty",
        }
    }
    if age < 0 || age > 150 {
        return &ValidationError{
            Field:   "age",
            Message: "must be between 0 and 150",
        }
    }
    return nil
}

func main() {
    err := validateUser("", 25)
    if err != nil {
        fmt.Println("Error:", err)

        // 类型断言获取详细信息
        if ve, ok := err.(*ValidationError); ok {
            fmt.Println("Invalid field:", ve.Field)
        }
    }
}
```

### 错误包装 (Go 1.13+)

```go
package main

import (
    "errors"
    "fmt"
    "os"
)

func readConfig(filename string) ([]byte, error) {
    data, err := os.ReadFile(filename)
    if err != nil {
        // 包装错误，添加上下文
        return nil, fmt.Errorf("failed to read config file: %w", err)
    }
    return data, nil
}

func main() {
    _, err := readConfig("config.json")
    if err != nil {
        fmt.Println("Error:", err)

        // 检查是否包含特定错误
        if errors.Is(err, os.ErrNotExist) {
            fmt.Println("Config file does not exist")
        }

        // 获取原始错误
        var pathErr *os.PathError
        if errors.As(err, &pathErr) {
            fmt.Println("Path:", pathErr.Path)
        }
    }
}
```

---

## 错误处理模式

### 模式 1: 早返回

```go
func processFile(filename string) error {
    file, err := os.Open(filename)
    if err != nil {
        return fmt.Errorf("open file: %w", err)
    }
    defer file.Close()

    data, err := io.ReadAll(file)
    if err != nil {
        return fmt.Errorf("read file: %w", err)
    }

    if len(data) == 0 {
        return errors.New("file is empty")
    }

    // 处理数据...
    return nil
}
```

### 模式 2: 错误变量

```go
package main

import (
    "errors"
    "fmt"
)

// 预定义错误
var (
    ErrNotFound     = errors.New("not found")
    ErrUnauthorized = errors.New("unauthorized")
    ErrInvalidInput = errors.New("invalid input")
)

func getResource(id int) (string, error) {
    if id <= 0 {
        return "", ErrInvalidInput
    }
    if id == 404 {
        return "", ErrNotFound
    }
    return fmt.Sprintf("Resource-%d", id), nil
}

func main() {
    _, err := getResource(404)
    if err != nil {
        switch {
        case errors.Is(err, ErrNotFound):
            fmt.Println("Resource not found")
        case errors.Is(err, ErrInvalidInput):
            fmt.Println("Invalid input")
        default:
            fmt.Println("Unknown error:", err)
        }
    }
}
```

### 模式 3: 错误处理函数

```go
package main

import (
    "fmt"
    "log"
)

func must[T any](val T, err error) T {
    if err != nil {
        log.Fatal(err)
    }
    return val
}

func mustParse(s string) int {
    var n int
    _, err := fmt.Sscanf(s, "%d", &n)
    if err != nil {
        log.Fatalf("failed to parse %q: %v", s, err)
    }
    return n
}

func main() {
    // 仅在确定不会出错或允许程序终止时使用
    // 不适合生产代码中的用户输入处理
}
```

---

## panic 和 recover

### panic

`panic` 用于不可恢复的错误，会终止程序执行。

```go
package main

import "fmt"

func main() {
    fmt.Println("Start")
    panic("something went wrong")
    fmt.Println("This won't be printed")
}
```

### 何时使用 panic

```go
// 1. 程序初始化失败
func init() {
    if os.Getenv("REQUIRED_CONFIG") == "" {
        panic("REQUIRED_CONFIG environment variable is not set")
    }
}

// 2. 程序员错误（不应该发生的情况）
func divide(a, b int) int {
    if b == 0 {
        panic("divide by zero: this should never happen")
    }
    return a / b
}

// 3. 实现接口的必须方法
func (m *MyType) MustImplement() {
    panic("not implemented")
}
```

### recover

`recover` 用于捕获 panic，只能在 defer 函数中使用。

```go
package main

import "fmt"

func mayPanic() {
    panic("a problem occurred")
}

func safeCall() {
    defer func() {
        if r := recover(); r != nil {
            fmt.Println("Recovered from:", r)
        }
    }()

    mayPanic()
    fmt.Println("This won't be printed")
}

func main() {
    safeCall()
    fmt.Println("Program continues...")
}
```

输出:
```
Recovered from: a problem occurred
Program continues...
```

### HTTP 服务器中的 recover

```go
package main

import (
    "fmt"
    "net/http"
)

func recoverMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                fmt.Println("Recovered from panic:", err)
                http.Error(w, "Internal Server Error", 500)
            }
        }()
        next(w, r)
    }
}

func riskyHandler(w http.ResponseWriter, r *http.Request) {
    panic("something went wrong")
}

func main() {
    http.HandleFunc("/", recoverMiddleware(riskyHandler))
    http.ListenAndServe(":8080", nil)
}
```

---

## 与 JavaScript/TypeScript 对比

### 异常处理对比

```javascript
// JavaScript - try/catch
function divide(a, b) {
    if (b === 0) {
        throw new Error("Division by zero");
    }
    return a / b;
}

try {
    const result = divide(10, 0);
    console.log(result);
} catch (err) {
    console.log("Error:", err.message);
}

// 异步错误
async function fetchData() {
    try {
        const response = await fetch(url);
        return await response.json();
    } catch (err) {
        console.log("Fetch error:", err);
    }
}
```

```go
// Go - 返回 error
func divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, errors.New("division by zero")
    }
    return a / b, nil
}

result, err := divide(10, 0)
if err != nil {
    fmt.Println("Error:", err)
} else {
    fmt.Println(result)
}
```

### 自定义错误对比

```javascript
// JavaScript
class ValidationError extends Error {
    constructor(field, message) {
        super(`${field}: ${message}`);
        this.field = field;
        this.name = "ValidationError";
    }
}

try {
    throw new ValidationError("email", "invalid format");
} catch (err) {
    if (err instanceof ValidationError) {
        console.log("Field:", err.field);
    }
}
```

```go
// Go
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

err := &ValidationError{Field: "email", Message: "invalid format"}
if ve, ok := err.(*ValidationError); ok {
    fmt.Println("Field:", ve.Field)
}
```

### 差异总结

| 特性 | JavaScript | Go |
|-----|-----------|-----|
| 错误机制 | 异常 (throw/catch) | 返回值 (error) |
| 错误传播 | 自动冒泡 | 显式返回 |
| 错误类型 | Error 类及子类 | error 接口 |
| 捕获语法 | try/catch/finally | if err != nil |
| 不可恢复错误 | 不区分 | panic/recover |
| 错误包装 | Error cause | fmt.Errorf %w |
| 空值检查 | null/undefined | nil |

### 错误处理风格对比

```javascript
// JavaScript - 异常风格
// 错误会自动传播，不需要每次检查
async function processUser(id) {
    const user = await getUser(id);       // 可能抛出
    const data = await fetchData(user);   // 可能抛出
    return transform(data);               // 可能抛出
}

// 调用时统一处理
try {
    const result = await processUser(123);
} catch (err) {
    console.log("Something went wrong:", err);
}
```

```go
// Go - 显式错误处理风格
func processUser(id int) (*Result, error) {
    user, err := getUser(id)
    if err != nil {
        return nil, fmt.Errorf("get user: %w", err)
    }

    data, err := fetchData(user)
    if err != nil {
        return nil, fmt.Errorf("fetch data: %w", err)
    }

    result, err := transform(data)
    if err != nil {
        return nil, fmt.Errorf("transform: %w", err)
    }

    return result, nil
}

// 调用
result, err := processUser(123)
if err != nil {
    log.Println("Error:", err)
    return
}
```

### 优缺点对比

| 方面 | JavaScript 异常 | Go 错误返回 |
|-----|----------------|------------|
| 代码简洁性 | 更简洁 | 更冗长 |
| 错误可见性 | 隐式 | 显式 |
| 错误处理强制性 | 可忽略 | 编译器提醒 |
| 控制流清晰度 | 可能跳转 | 线性 |
| 性能 | 异常开销大 | 几乎无开销 |

---

## 最佳实践

### 1. 总是检查错误

```go
// 不好
result, _ := doSomething()

// 好
result, err := doSomething()
if err != nil {
    return err
}
```

### 2. 添加上下文

```go
// 不好
if err != nil {
    return err
}

// 好
if err != nil {
    return fmt.Errorf("processing user %d: %w", userID, err)
}
```

### 3. 使用 errors.Is 和 errors.As

```go
// 检查特定错误
if errors.Is(err, os.ErrNotExist) {
    // 文件不存在
}

// 获取特定错误类型
var pathErr *os.PathError
if errors.As(err, &pathErr) {
    fmt.Println("Path:", pathErr.Path)
}
```

### 4. 预定义可导出的错误

```go
package mypackage

var (
    ErrNotFound = errors.New("not found")
    ErrInvalid  = errors.New("invalid")
)
```

### 5. panic 仅用于程序员错误

```go
// 好 - 真正不应该发生的情况
switch v := value.(type) {
case int:
    // ...
case string:
    // ...
default:
    panic(fmt.Sprintf("unexpected type: %T", v))
}

// 不好 - 用户输入错误不应该 panic
func parseInput(s string) int {
    n, err := strconv.Atoi(s)
    if err != nil {
        panic(err)  // 不好！应该返回错误
    }
    return n
}
```

---

## 下一步

- [包管理](./08-packages.md) - 学习 Go 的包和模块系统
