# 类型断言

本文档介绍 Go 语言中的类型断言和类型开关。

## 目录

- [类型断言基础](#类型断言基础)
- [安全的类型断言](#安全的类型断言)
- [类型开关](#类型开关)
- [接口值的内部结构](#接口值的内部结构)
- [常见使用场景](#常见使用场景)
- [与 JavaScript/TypeScript 对比](#与-javascripttypescript-对比)

---

## 类型断言基础

### 什么是类型断言？

类型断言用于从接口值中提取其底层的具体类型。

```go
package main

import "fmt"

func main() {
    var i interface{} = "hello"

    // 类型断言：i.(T)
    s := i.(string)
    fmt.Println(s) // hello

    // 断言为错误类型会 panic
    // n := i.(int)  // panic: interface conversion
}
```

### 语法

```go
value := interfaceValue.(ConcreteType)
```

- `interfaceValue`: 接口类型的变量
- `ConcreteType`: 要断言的具体类型
- `value`: 断言成功后的具体类型值

### 断言失败会 Panic

```go
func main() {
    var i interface{} = "hello"

    // 这会 panic，因为 i 不是 int 类型
    n := i.(int)
    fmt.Println(n)
}
// panic: interface conversion: interface {} is string, not int
```

---

## 安全的类型断言

### 使用 comma-ok 模式

```go
func main() {
    var i interface{} = "hello"

    // 安全的类型断言：返回两个值
    s, ok := i.(string)
    if ok {
        fmt.Println("String value:", s)
    }

    // 断言失败时，ok 为 false，n 为零值
    n, ok := i.(int)
    if !ok {
        fmt.Println("Not an int, got zero value:", n) // n = 0
    }
}
```

### 类型断言的返回值

```go
value, ok := i.(Type)
// value: 如果断言成功，是具体类型的值；否则是类型的零值
// ok: 布尔值，表示断言是否成功
```

### 实际应用

```go
func ProcessValue(v interface{}) {
    // 安全地处理不同类型
    if s, ok := v.(string); ok {
        fmt.Println("Processing string:", s)
        return
    }

    if n, ok := v.(int); ok {
        fmt.Println("Processing int:", n)
        return
    }

    fmt.Println("Unknown type")
}

func main() {
    ProcessValue("hello")  // Processing string: hello
    ProcessValue(42)       // Processing int: 42
    ProcessValue(3.14)     // Unknown type
}
```

---

## 类型开关

### 基本语法

类型开关是处理多种类型的更优雅方式：

```go
func describe(i interface{}) {
    switch v := i.(type) {
    case int:
        fmt.Printf("Integer: %d\n", v)
    case string:
        fmt.Printf("String: %s\n", v)
    case bool:
        fmt.Printf("Boolean: %t\n", v)
    case []int:
        fmt.Printf("Int slice with %d elements\n", len(v))
    default:
        fmt.Printf("Unknown type: %T\n", v)
    }
}

func main() {
    describe(42)           // Integer: 42
    describe("hello")      // String: hello
    describe(true)         // Boolean: true
    describe([]int{1,2,3}) // Int slice with 3 elements
    describe(3.14)         // Unknown type: float64
}
```

### 多类型匹配

```go
func printValue(v interface{}) {
    switch v.(type) {
    case int, int8, int16, int32, int64:
        fmt.Println("This is an integer type")
    case uint, uint8, uint16, uint32, uint64:
        fmt.Println("This is an unsigned integer type")
    case float32, float64:
        fmt.Println("This is a float type")
    case string:
        fmt.Println("This is a string")
    case bool:
        fmt.Println("This is a boolean")
    default:
        fmt.Println("Unknown type")
    }
}
```

### 类型开关中使用变量

```go
func process(i interface{}) string {
    switch v := i.(type) {
    case string:
        return "String of length " + fmt.Sprint(len(v))
    case int:
        return "Integer squared: " + fmt.Sprint(v*v)
    case nil:
        return "Nil value"
    default:
        return fmt.Sprintf("Unhandled type: %T", v)
    }
}
```

### 断言到接口

```go
type Stringer interface {
    String() string
}

type Numbered interface {
    Number() int
}

func describe(i interface{}) {
    switch v := i.(type) {
    case Stringer:
        fmt.Println("Stringer:", v.String())
    case Numbered:
        fmt.Println("Numbered:", v.Number())
    default:
        fmt.Println("Neither Stringer nor Numbered")
    }
}
```

---

## 接口值的内部结构

### 接口值的组成

每个接口值由两部分组成：
1. **类型信息**：底层具体类型
2. **值信息**：底层具体值

```go
func main() {
    var i interface{}

    // 空接口值：类型和值都是 nil
    fmt.Printf("Type: %T, Value: %v\n", i, i)
    // Type: <nil>, Value: <nil>

    i = 42
    fmt.Printf("Type: %T, Value: %v\n", i, i)
    // Type: int, Value: 42

    i = "hello"
    fmt.Printf("Type: %T, Value: %v\n", i, i)
    // Type: string, Value: hello
}
```

### nil 接口值 vs 持有 nil 的接口值

```go
type MyError struct {
    Message string
}

func (e *MyError) Error() string {
    return e.Message
}

func returnsError(fail bool) error {
    var err *MyError = nil
    if fail {
        err = &MyError{Message: "something went wrong"}
    }
    return err  // 注意：即使 err 是 nil，返回的接口值不是 nil！
}

func main() {
    err := returnsError(false)

    // 这个检查会失败！
    if err != nil {
        fmt.Println("Error is not nil!") // 会打印这个
        fmt.Printf("Type: %T, Value: %v\n", err, err)
        // Type: *main.MyError, Value: <nil>
    }
}
```

### 正确的 nil 返回

```go
func returnsErrorCorrectly(fail bool) error {
    if fail {
        return &MyError{Message: "something went wrong"}
    }
    return nil  // 直接返回 nil，而不是 nil 指针
}

func main() {
    err := returnsErrorCorrectly(false)
    if err == nil {
        fmt.Println("No error") // 正确：No error
    }
}
```

---

## 常见使用场景

### 1. 处理 JSON 数据

```go
import "encoding/json"

func processJSON(data []byte) {
    var result interface{}
    json.Unmarshal(data, &result)

    // 使用类型开关处理不同 JSON 结构
    switch v := result.(type) {
    case map[string]interface{}:
        fmt.Println("JSON Object with keys:", len(v))
        for key, value := range v {
            fmt.Printf("  %s: %v\n", key, value)
        }
    case []interface{}:
        fmt.Println("JSON Array with", len(v), "elements")
    case string:
        fmt.Println("JSON String:", v)
    case float64:
        fmt.Println("JSON Number:", v)
    case bool:
        fmt.Println("JSON Boolean:", v)
    case nil:
        fmt.Println("JSON null")
    }
}
```

### 2. 错误类型检查

```go
import (
    "errors"
    "os"
)

func handleError(err error) {
    // 检查特定错误类型
    var pathErr *os.PathError
    if errors.As(err, &pathErr) {
        fmt.Printf("Path error on %s: %v\n", pathErr.Path, pathErr.Err)
        return
    }

    // 使用类型开关
    switch e := err.(type) {
    case *os.PathError:
        fmt.Println("Path error:", e.Path)
    case *os.LinkError:
        fmt.Println("Link error:", e.Old, "->", e.New)
    default:
        fmt.Println("Other error:", err)
    }
}
```

### 3. 实现多态行为

```go
type Shape interface {
    Area() float64
}

type Rectangle struct {
    Width, Height float64
}

func (r Rectangle) Area() float64 {
    return r.Width * r.Height
}

type Circle struct {
    Radius float64
}

func (c Circle) Area() float64 {
    return 3.14159 * c.Radius * c.Radius
}

func describeShape(s Shape) {
    fmt.Printf("Area: %.2f", s.Area())

    // 获取特定类型的额外信息
    switch shape := s.(type) {
    case Rectangle:
        fmt.Printf(" (Rectangle: %.2f x %.2f)\n", shape.Width, shape.Height)
    case Circle:
        fmt.Printf(" (Circle: radius %.2f)\n", shape.Radius)
    default:
        fmt.Println(" (Unknown shape)")
    }
}

func main() {
    describeShape(Rectangle{Width: 10, Height: 5})
    // Area: 50.00 (Rectangle: 10.00 x 5.00)

    describeShape(Circle{Radius: 3})
    // Area: 28.27 (Circle: radius 3.00)
}
```

### 4. 类型断言用于方法调用

```go
type Closable interface {
    Close() error
}

func MaybeClose(v interface{}) error {
    if c, ok := v.(Closable); ok {
        return c.Close()
    }
    return nil
}

// 或使用 io.Closer
import "io"

func CloseIfPossible(v interface{}) error {
    if closer, ok := v.(io.Closer); ok {
        return closer.Close()
    }
    return nil
}
```

---

## 与 JavaScript/TypeScript 对比

### 类型检查对比

```typescript
// TypeScript - typeof 和 instanceof
function process(value: unknown) {
    if (typeof value === "string") {
        console.log(value.toUpperCase());
    } else if (typeof value === "number") {
        console.log(value * 2);
    } else if (value instanceof Array) {
        console.log(value.length);
    }
}
```

```go
// Go - 类型断言和类型开关
func process(value interface{}) {
    switch v := value.(type) {
    case string:
        fmt.Println(strings.ToUpper(v))
    case int:
        fmt.Println(v * 2)
    case []interface{}:
        fmt.Println(len(v))
    }
}
```

### 类型守卫对比

```typescript
// TypeScript - 类型守卫
interface Dog {
    bark(): void;
}

interface Cat {
    meow(): void;
}

function isDog(animal: Dog | Cat): animal is Dog {
    return (animal as Dog).bark !== undefined;
}

function makeSound(animal: Dog | Cat) {
    if (isDog(animal)) {
        animal.bark();  // TypeScript 知道这是 Dog
    } else {
        animal.meow();  // TypeScript 知道这是 Cat
    }
}
```

```go
// Go - 类型断言
type Dog interface {
    Bark()
}

type Cat interface {
    Meow()
}

type Animal interface{}

func MakeSound(animal Animal) {
    if dog, ok := animal.(Dog); ok {
        dog.Bark()
    } else if cat, ok := animal.(Cat); ok {
        cat.Meow()
    }
}
```

### 联合类型对比

```typescript
// TypeScript - 联合类型
type StringOrNumber = string | number;

function double(value: StringOrNumber): StringOrNumber {
    if (typeof value === "string") {
        return value + value;
    }
    return value * 2;
}
```

```go
// Go - 使用接口和类型断言
func double(value interface{}) interface{} {
    switch v := value.(type) {
    case string:
        return v + v
    case int:
        return v * 2
    case float64:
        return v * 2
    default:
        return nil
    }
}

// Go 1.18+ 可以使用泛型约束
type Doubleable interface {
    string | int | float64
}

func doubleGeneric[T Doubleable](value T) T {
    // 仍然需要类型断言处理不同逻辑
    var result interface{} = value
    switch v := result.(type) {
    case string:
        return interface{}(v + v).(T)
    default:
        // 数值类型
        return value + value
    }
}
```

### 主要差异

| 特性 | TypeScript | Go |
|-----|-----------|-----|
| 类型检查 | `typeof`, `instanceof` | 类型断言 `.(Type)` |
| 类型守卫 | 函数返回 `x is Type` | comma-ok 模式 |
| 多类型处理 | `switch (typeof x)` | `switch x.(type)` |
| 失败处理 | 编译时类型收窄 | 运行时 panic 或 ok=false |
| 联合类型 | `Type1 \| Type2` | 接口 + 类型断言 |

---

## 最佳实践

### 1. 始终使用 comma-ok 模式

```go
// 好
if v, ok := i.(string); ok {
    // 使用 v
}

// 危险 - 可能 panic
v := i.(string)
```

### 2. 优先使用类型开关

```go
// 好 - 清晰的多类型处理
switch v := i.(type) {
case int:
    // ...
case string:
    // ...
}

// 避免 - 多个 if-else
if _, ok := i.(int); ok {
    // ...
} else if _, ok := i.(string); ok {
    // ...
}
```

### 3. 限制空接口的使用

```go
// 避免
func Process(data interface{}) {}

// 推荐 - 使用具体类型或定义接口
type Processor interface {
    Process()
}

func Process(p Processor) {}
```

---

## 下一步

- [嵌入与组合](./06-embedding.md) - 学习结构体嵌入实现代码复用
