# 方法

本文档介绍 Go 语言中方法的定义，特别是值接收者和指针接收者的区别。

## 目录

- [方法基础](#方法基础)
- [值接收者 vs 指针接收者](#值接收者-vs-指针接收者)
- [方法值与方法表达式](#方法值与方法表达式)
- [方法与函数的区别](#方法与函数的区别)
- [与 JavaScript/TypeScript 对比](#与-javascripttypescript-对比)

---

## 方法基础

### 什么是方法？

方法是带有接收者的函数。接收者指定了方法属于哪个类型。

```go
package main

import "fmt"

type Rectangle struct {
    Width  float64
    Height float64
}

// 定义方法：func (接收者) 方法名(参数) 返回值
func (r Rectangle) Area() float64 {
    return r.Width * r.Height
}

func (r Rectangle) Perimeter() float64 {
    return 2 * (r.Width + r.Height)
}

func main() {
    rect := Rectangle{Width: 10, Height: 5}

    fmt.Println("Area:", rect.Area())           // 50
    fmt.Println("Perimeter:", rect.Perimeter()) // 30
}
```

### 方法语法

```go
func (receiver Type) MethodName(params) returnType {
    // 方法体
}
```

- `receiver`: 接收者变量名（通常使用类型首字母小写）
- `Type`: 接收者类型
- 接收者类似于其他语言中的 `this` 或 `self`

### 接收者命名约定

```go
// 好的命名
func (r Rectangle) Area() float64 { ... }
func (c Circle) Area() float64 { ... }
func (u *User) Save() error { ... }

// 不推荐
func (this Rectangle) Area() float64 { ... }
func (self Circle) Area() float64 { ... }
func (rectangle Rectangle) Area() float64 { ... }
```

---

## 值接收者 vs 指针接收者

### 值接收者

值接收者操作的是结构体的副本，不能修改原结构体。

```go
type Counter struct {
    Value int
}

// 值接收者
func (c Counter) Increment() {
    c.Value++  // 只修改副本
}

func (c Counter) GetValue() int {
    return c.Value
}

func main() {
    counter := Counter{Value: 0}
    counter.Increment()
    fmt.Println(counter.Value)  // 0 - 原值未改变
}
```

### 指针接收者

指针接收者可以修改原结构体。

```go
// 指针接收者
func (c *Counter) IncrementPointer() {
    c.Value++  // 修改原对象
}

func main() {
    counter := Counter{Value: 0}
    counter.IncrementPointer()
    fmt.Println(counter.Value)  // 1 - 原值已改变
}
```

### 自动转换

Go 会自动在值和指针之间转换：

```go
func main() {
    // 值调用指针方法
    counter := Counter{Value: 0}
    counter.IncrementPointer()  // 自动转换为 (&counter).IncrementPointer()

    // 指针调用值方法
    counterPtr := &Counter{Value: 10}
    value := counterPtr.GetValue()  // 自动转换为 (*counterPtr).GetValue()
}
```

### 选择接收者类型

| 使用指针接收者 | 使用值接收者 |
|--------------|------------|
| 需要修改接收者 | 不需要修改接收者 |
| 接收者是大结构体 | 接收者是小结构体（如 Point） |
| 保持一致性（其他方法用指针） | 基本类型的别名 |
| 接收者包含同步原语（如 Mutex） | |

### 完整示例

```go
package main

import "fmt"

type User struct {
    Name  string
    Email string
    Age   int
}

// 值接收者 - 只读操作
func (u User) String() string {
    return fmt.Sprintf("%s <%s>", u.Name, u.Email)
}

func (u User) IsAdult() bool {
    return u.Age >= 18
}

// 指针接收者 - 修改操作
func (u *User) SetName(name string) {
    u.Name = name
}

func (u *User) HaveBirthday() {
    u.Age++
}

// 指针接收者 - 返回错误的操作
func (u *User) UpdateEmail(email string) error {
    if email == "" {
        return fmt.Errorf("email cannot be empty")
    }
    u.Email = email
    return nil
}

func main() {
    user := User{Name: "Alice", Email: "alice@example.com", Age: 17}

    fmt.Println(user.String())      // Alice <alice@example.com>
    fmt.Println(user.IsAdult())     // false

    user.HaveBirthday()
    fmt.Println(user.IsAdult())     // true

    user.SetName("Alice Smith")
    fmt.Println(user.String())      // Alice Smith <alice@example.com>
}
```

---

## 方法值与方法表达式

### 方法值

方法可以赋值给变量：

```go
type Adder struct {
    Value int
}

func (a Adder) Add(n int) int {
    return a.Value + n
}

func main() {
    adder := Adder{Value: 10}

    // 方法值 - 绑定了接收者
    addFunc := adder.Add
    fmt.Println(addFunc(5))  // 15
    fmt.Println(addFunc(10)) // 20
}
```

### 方法表达式

方法表达式需要显式传递接收者：

```go
func main() {
    // 方法表达式 - 需要传递接收者
    addExpr := Adder.Add

    adder1 := Adder{Value: 10}
    adder2 := Adder{Value: 20}

    fmt.Println(addExpr(adder1, 5))  // 15
    fmt.Println(addExpr(adder2, 5))  // 25
}
```

### 方法作为回调

```go
type Button struct {
    Label   string
    OnClick func()
}

type Handler struct {
    Message string
}

func (h *Handler) HandleClick() {
    fmt.Println("Clicked:", h.Message)
}

func main() {
    handler := &Handler{Message: "Hello"}

    button := Button{
        Label:   "Click me",
        OnClick: handler.HandleClick,  // 方法值作为回调
    }

    button.OnClick()  // Clicked: Hello
}
```

---

## 方法与函数的区别

### 定义方式

```go
// 函数
func AreaFunc(r Rectangle) float64 {
    return r.Width * r.Height
}

// 方法
func (r Rectangle) AreaMethod() float64 {
    return r.Width * r.Height
}
```

### 调用方式

```go
rect := Rectangle{Width: 10, Height: 5}

// 函数调用
area1 := AreaFunc(rect)

// 方法调用
area2 := rect.AreaMethod()
```

### 为非结构体定义方法

可以为任何自定义类型定义方法：

```go
// 自定义类型
type MyInt int

func (m MyInt) Double() MyInt {
    return m * 2
}

func (m MyInt) IsPositive() bool {
    return m > 0
}

func main() {
    num := MyInt(5)
    fmt.Println(num.Double())      // 10
    fmt.Println(num.IsPositive())  // true
}
```

### 不能为内置类型直接定义方法

```go
// 错误：不能为 int 定义方法
// func (i int) Double() int { return i * 2 }

// 正确：创建类型别名
type MyInt int
func (m MyInt) Double() MyInt { return m * 2 }
```

---

## 与 JavaScript/TypeScript 对比

### 方法定义对比

```typescript
// TypeScript
class Rectangle {
    constructor(
        private width: number,
        private height: number
    ) {}

    area(): number {
        return this.width * this.height;
    }

    // 静态方法
    static createSquare(size: number): Rectangle {
        return new Rectangle(size, size);
    }
}
```

```go
// Go
type Rectangle struct {
    Width  float64
    Height float64
}

func (r Rectangle) Area() float64 {
    return r.Width * r.Height
}

// Go 没有静态方法，使用普通函数
func CreateSquare(size float64) Rectangle {
    return Rectangle{Width: size, Height: size}
}
```

### this vs 接收者

```typescript
// TypeScript - this
class Counter {
    private value = 0;

    increment() {
        this.value++;  // 使用 this
    }

    getValue(): number {
        return this.value;
    }
}
```

```go
// Go - 接收者
type Counter struct {
    value int
}

func (c *Counter) Increment() {
    c.value++  // 使用接收者变量名
}

func (c Counter) GetValue() int {
    return c.value
}
```

### 链式调用

```typescript
// TypeScript
class Builder {
    private value = "";

    append(s: string): this {
        this.value += s;
        return this;  // 返回 this
    }

    build(): string {
        return this.value;
    }
}

const result = new Builder()
    .append("Hello")
    .append(" ")
    .append("World")
    .build();
```

```go
// Go
type Builder struct {
    value string
}

func (b *Builder) Append(s string) *Builder {
    b.value += s
    return b  // 返回指针
}

func (b *Builder) Build() string {
    return b.value
}

func main() {
    result := new(Builder).
        Append("Hello").
        Append(" ").
        Append("World").
        Build()
}
```

### 主要差异

| 特性 | TypeScript | Go |
|-----|-----------|-----|
| this/self | `this` | 接收者变量 |
| 方法定义 | 类内部 | 类外部 |
| 静态方法 | `static` | 普通函数 |
| 方法可见性 | `public/private` | 首字母大小写 |
| 值/引用 | 对象总是引用 | 值/指针接收者 |

---

## 最佳实践

### 1. 一致的接收者类型

```go
// 好 - 所有方法使用相同类型的接收者
type User struct { ... }

func (u *User) SetName(name string) { ... }
func (u *User) SetEmail(email string) { ... }
func (u *User) Save() error { ... }
func (u *User) String() string { ... }  // 即使不需要修改，也用指针保持一致
```

### 2. 小结构体可以使用值接收者

```go
type Point struct {
    X, Y int
}

// 小结构体使用值接收者是可以的
func (p Point) Distance(other Point) float64 {
    dx := float64(p.X - other.X)
    dy := float64(p.Y - other.Y)
    return math.Sqrt(dx*dx + dy*dy)
}
```

### 3. 避免在方法中使用 nil 接收者

```go
func (u *User) Name() string {
    if u == nil {
        return ""  // 防御性编程
    }
    return u.name
}
```

### 4. Getter/Setter 命名

```go
// Go 风格
func (u *User) Name() string { return u.name }
func (u *User) SetName(name string) { u.name = name }

// 不是 Go 风格
func (u *User) GetName() string { ... }  // 不需要 Get 前缀
```

---

## 下一步

- [接口](./04-interfaces.md) - 学习 Go 的接口系统
