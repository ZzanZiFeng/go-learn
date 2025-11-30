# 嵌入与组合

本文档介绍 Go 语言的结构体嵌入机制，实现"组合优于继承"的设计理念。

## 目录

- [组合 vs 继承](#组合-vs-继承)
- [结构体嵌入](#结构体嵌入)
- [接口嵌入](#接口嵌入)
- [方法提升](#方法提升)
- [嵌入的应用模式](#嵌入的应用模式)
- [与 JavaScript/TypeScript 对比](#与-javascripttypescript-对比)

---

## 组合 vs 继承

### Go 的设计哲学

Go 没有传统的类继承，而是通过**组合**来复用代码。这被称为"组合优于继承"原则。

| 继承 | 组合 |
|-----|-----|
| "is-a" 关系 | "has-a" 关系 |
| 紧耦合 | 松耦合 |
| 难以修改基类 | 灵活组合 |
| 菱形继承问题 | 无此问题 |

### 组合的优势

```go
// 继承思维（Go 不支持）
// type Dog extends Animal { ... }

// 组合思维（Go 推荐）
type Animal struct {
    Name string
    Age  int
}

type Dog struct {
    Animal  // 嵌入 Animal
    Breed   string
}

// Dog "有一个" Animal，而不是 Dog "是一个" Animal
```

---

## 结构体嵌入

### 基本嵌入

```go
package main

import "fmt"

type Address struct {
    City    string
    Country string
}

type Person struct {
    Name    string
    Age     int
    Address // 匿名嵌入（没有字段名）
}

func main() {
    p := Person{
        Name: "Alice",
        Age:  30,
        Address: Address{
            City:    "Beijing",
            Country: "China",
        },
    }

    // 直接访问嵌入字段的属性
    fmt.Println(p.City)    // Beijing（提升的字段）
    fmt.Println(p.Country) // China

    // 也可以通过类型名访问
    fmt.Println(p.Address.City) // Beijing
}
```

### 命名嵌入 vs 匿名嵌入

```go
// 匿名嵌入 - 字段会被提升
type Employee struct {
    Person  // 匿名
    Title   string
}

// 命名嵌入 - 字段不会被提升
type Employee2 struct {
    person Person // 命名字段
    Title  string
}

func main() {
    e1 := Employee{
        Person: Person{Name: "Bob", Age: 25},
        Title:  "Engineer",
    }
    fmt.Println(e1.Name)  // 直接访问（提升）

    e2 := Employee2{
        person: Person{Name: "Bob", Age: 25},
        Title:  "Engineer",
    }
    fmt.Println(e2.person.Name)  // 必须通过字段名访问
}
```

### 多层嵌入

```go
type Base struct {
    ID int
}

type Middle struct {
    Base
    Name string
}

type Top struct {
    Middle
    Value string
}

func main() {
    t := Top{
        Middle: Middle{
            Base: Base{ID: 1},
            Name: "test",
        },
        Value: "hello",
    }

    // 所有字段都被提升
    fmt.Println(t.ID)    // 1
    fmt.Println(t.Name)  // test
    fmt.Println(t.Value) // hello
}
```

### 字段名冲突

```go
type A struct {
    Name string
}

type B struct {
    Name string
}

type C struct {
    A
    B
    // Name string  // 如果 C 也有 Name，会覆盖 A 和 B 的
}

func main() {
    c := C{
        A: A{Name: "from A"},
        B: B{Name: "from B"},
    }

    // c.Name  // 编译错误：ambiguous selector c.Name

    // 必须明确指定
    fmt.Println(c.A.Name) // from A
    fmt.Println(c.B.Name) // from B
}
```

---

## 接口嵌入

### 在结构体中嵌入接口

```go
type Reader interface {
    Read(p []byte) (n int, err error)
}

type MyReader struct {
    Reader  // 嵌入接口
    source  string
}

// 如果没有实现 Read，MyReader 仍然满足 Reader 接口
// 但调用时会 panic（nil 接口）

func main() {
    var r Reader = &MyReader{}
    // r.Read(nil)  // panic: nil pointer dereference
}
```

### 接口嵌入的实际用途

```go
import (
    "io"
    "strings"
)

type CountingReader struct {
    io.Reader  // 嵌入 io.Reader 接口
    BytesRead  int
}

func (c *CountingReader) Read(p []byte) (n int, err error) {
    n, err = c.Reader.Read(p)  // 调用被嵌入的 Reader
    c.BytesRead += n
    return
}

func main() {
    sr := strings.NewReader("Hello, World!")
    cr := &CountingReader{Reader: sr}

    buf := make([]byte, 5)
    cr.Read(buf)
    fmt.Println(string(buf))      // Hello
    fmt.Println(cr.BytesRead)     // 5
}
```

---

## 方法提升

### 基本方法提升

```go
type Engine struct {
    Power int
}

func (e Engine) Start() {
    fmt.Println("Engine starting with power:", e.Power)
}

func (e *Engine) Stop() {
    fmt.Println("Engine stopping")
}

type Car struct {
    Engine  // 嵌入
    Brand   string
}

func main() {
    car := Car{
        Engine: Engine{Power: 200},
        Brand:  "Tesla",
    }

    // 方法被提升
    car.Start()  // Engine starting with power: 200
    car.Stop()   // Engine stopping

    // 也可以显式调用
    car.Engine.Start()
}
```

### 方法覆盖

```go
type Animal struct {
    Name string
}

func (a Animal) Speak() string {
    return "..."
}

type Dog struct {
    Animal
}

// Dog 覆盖 Speak 方法
func (d Dog) Speak() string {
    return d.Name + " says: Woof!"
}

func main() {
    dog := Dog{Animal: Animal{Name: "Buddy"}}

    fmt.Println(dog.Speak())        // Buddy says: Woof!
    fmt.Println(dog.Animal.Speak()) // ...（调用原始方法）
}
```

### 指针嵌入

```go
type Counter struct {
    count int
}

func (c *Counter) Increment() {
    c.count++
}

func (c *Counter) Value() int {
    return c.count
}

type Button struct {
    *Counter  // 嵌入指针
    Label     string
}

func main() {
    counter := &Counter{}
    btn := Button{
        Counter: counter,
        Label:   "Click me",
    }

    btn.Increment()
    btn.Increment()
    fmt.Println(btn.Value())       // 2
    fmt.Println(counter.Value())   // 2（同一个 Counter）
}
```

---

## 嵌入的应用模式

### 1. 混入模式 (Mixin)

```go
// 日志混入
type Logger struct{}

func (l Logger) Log(msg string) {
    fmt.Printf("[LOG] %s\n", msg)
}

func (l Logger) Error(msg string) {
    fmt.Printf("[ERROR] %s\n", msg)
}

// 指标混入
type Metrics struct {
    requestCount int
}

func (m *Metrics) IncrementRequests() {
    m.requestCount++
}

func (m *Metrics) GetRequestCount() int {
    return m.requestCount
}

// 服务组合多个混入
type UserService struct {
    Logger
    *Metrics
}

func (s *UserService) CreateUser(name string) {
    s.Log("Creating user: " + name)
    s.IncrementRequests()
}

func main() {
    service := &UserService{
        Metrics: &Metrics{},
    }

    service.CreateUser("Alice")
    service.CreateUser("Bob")

    fmt.Println("Total requests:", service.GetRequestCount()) // 2
}
```

### 2. 装饰器模式

```go
import (
    "io"
    "time"
)

// 基础 Writer
type SlowWriter struct {
    delay time.Duration
}

func (w SlowWriter) Write(p []byte) (n int, err error) {
    time.Sleep(w.delay)
    return len(p), nil
}

// 装饰器：添加缓冲
type BufferedWriter struct {
    io.Writer
    buffer []byte
}

func (b *BufferedWriter) Write(p []byte) (n int, err error) {
    b.buffer = append(b.buffer, p...)
    if len(b.buffer) > 1024 {
        n, err = b.Writer.Write(b.buffer)
        b.buffer = nil
    }
    return len(p), nil
}

// 装饰器：添加计数
type CountingWriter struct {
    io.Writer
    Count int
}

func (c *CountingWriter) Write(p []byte) (n int, err error) {
    n, err = c.Writer.Write(p)
    c.Count += n
    return
}
```

### 3. 适配器模式

```go
// 旧接口
type OldPrinter interface {
    PrintLn(s string)
}

type LegacyPrinter struct{}

func (p LegacyPrinter) PrintLn(s string) {
    fmt.Println("[Legacy]", s)
}

// 新接口
type Printer interface {
    Print(s string)
}

// 适配器
type PrinterAdapter struct {
    OldPrinter
}

func (a PrinterAdapter) Print(s string) {
    a.PrintLn(s)  // 调用旧方法
}

func main() {
    var printer Printer = PrinterAdapter{
        OldPrinter: LegacyPrinter{},
    }
    printer.Print("Hello")  // [Legacy] Hello
}
```

### 4. 包装同步原语

```go
import "sync"

type SafeMap struct {
    sync.RWMutex
    data map[string]interface{}
}

func NewSafeMap() *SafeMap {
    return &SafeMap{
        data: make(map[string]interface{}),
    }
}

func (m *SafeMap) Get(key string) (interface{}, bool) {
    m.RLock()
    defer m.RUnlock()
    val, ok := m.data[key]
    return val, ok
}

func (m *SafeMap) Set(key string, value interface{}) {
    m.Lock()
    defer m.Unlock()
    m.data[key] = value
}

func main() {
    m := NewSafeMap()
    m.Set("name", "Alice")
    if val, ok := m.Get("name"); ok {
        fmt.Println(val)  // Alice
    }
}
```

---

## 与 JavaScript/TypeScript 对比

### 继承 vs 组合

```typescript
// TypeScript - 继承
class Animal {
    name: string;

    constructor(name: string) {
        this.name = name;
    }

    speak(): string {
        return "...";
    }
}

class Dog extends Animal {
    breed: string;

    constructor(name: string, breed: string) {
        super(name);  // 调用父类构造函数
        this.breed = breed;
    }

    speak(): string {
        return `${this.name} says: Woof!`;
    }
}
```

```go
// Go - 组合
type Animal struct {
    Name string
}

func (a Animal) Speak() string {
    return "..."
}

type Dog struct {
    Animal  // 嵌入
    Breed   string
}

func (d Dog) Speak() string {
    return d.Name + " says: Woof!"
}
```

### Mixin 对比

```typescript
// TypeScript - Mixin
type Constructor<T = {}> = new (...args: any[]) => T;

function Timestamped<TBase extends Constructor>(Base: TBase) {
    return class extends Base {
        createdAt = new Date();
    };
}

function Activatable<TBase extends Constructor>(Base: TBase) {
    return class extends Base {
        isActive = false;
        activate() { this.isActive = true; }
        deactivate() { this.isActive = false; }
    };
}

class User {}
const TimestampedActivatableUser = Timestamped(Activatable(User));
```

```go
// Go - Mixin 通过嵌入
type Timestamped struct {
    CreatedAt time.Time
}

type Activatable struct {
    IsActive bool
}

func (a *Activatable) Activate() {
    a.IsActive = true
}

func (a *Activatable) Deactivate() {
    a.IsActive = false
}

type User struct {
    Timestamped
    Activatable
    Name string
}

func main() {
    u := User{
        Timestamped: Timestamped{CreatedAt: time.Now()},
        Name:        "Alice",
    }
    u.Activate()
    fmt.Println(u.IsActive)   // true
    fmt.Println(u.CreatedAt)  // 当前时间
}
```

### super 调用对比

```typescript
// TypeScript
class Child extends Parent {
    method() {
        super.method();  // 调用父类方法
    }
}
```

```go
// Go - 显式调用嵌入类型的方法
type Child struct {
    Parent
}

func (c Child) Method() {
    c.Parent.Method()  // 调用嵌入类型的方法
}
```

### 主要差异

| 特性 | TypeScript | Go |
|-----|-----------|-----|
| 继承语法 | `extends` | 嵌入 |
| 调用父类 | `super.method()` | `embedded.Method()` |
| 构造函数继承 | 自动 + `super()` | 手动组合 |
| 多继承 | 不支持 | 多嵌入 |
| 方法覆盖 | `override` | 同名方法 |
| 访问控制 | `public/private` | 首字母大小写 |

---

## 最佳实践

### 1. 优先使用组合

```go
// 推荐：组合
type UserService struct {
    db       *Database
    cache    *Cache
    logger   *Logger
}

// 而不是：多层继承（Go 不支持）
```

### 2. 嵌入用于 "is-a" 关系

```go
// 好：File 确实是 ReadWriter
type File struct {
    io.ReadWriter
    path string
}

// 不好：User 不是 Logger
type User struct {
    Logger  // 应该用命名字段
    Name string
}
```

### 3. 避免过深嵌入

```go
// 避免
type A struct{ B }
type B struct{ C }
type C struct{ D }
type D struct{ E }

// 推荐：扁平化设计
type A struct {
    b B
    c C
    d D
    e E
}
```

---

## 下一步

- [泛型基础](./07-generics.md) - 学习 Go 1.18+ 泛型入门
