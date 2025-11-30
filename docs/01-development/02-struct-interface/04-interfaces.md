# 接口

本文档介绍 Go 语言接口的定义和隐式实现机制。

## 目录

- [接口基础](#接口基础)
- [隐式实现](#隐式实现)
- [空接口](#空接口)
- [接口组合](#接口组合)
- [常用标准接口](#常用标准接口)
- [接口最佳实践](#接口最佳实践)
- [与 JavaScript/TypeScript 对比](#与-javascripttypescript-对比)

---

## 接口基础

### 什么是接口？

接口定义了一组方法签名，任何实现了这些方法的类型都隐式实现了该接口。

```go
package main

import "fmt"

// 定义接口
type Speaker interface {
    Speak() string
}

// Dog 类型
type Dog struct {
    Name string
}

// Dog 实现 Speaker 接口（隐式）
func (d Dog) Speak() string {
    return d.Name + " says: Woof!"
}

// Cat 类型
type Cat struct {
    Name string
}

// Cat 实现 Speaker 接口（隐式）
func (c Cat) Speak() string {
    return c.Name + " says: Meow!"
}

func main() {
    // 接口变量可以持有任何实现该接口的值
    var speaker Speaker

    speaker = Dog{Name: "Buddy"}
    fmt.Println(speaker.Speak()) // Buddy says: Woof!

    speaker = Cat{Name: "Whiskers"}
    fmt.Println(speaker.Speak()) // Whiskers says: Meow!
}
```

### 接口语法

```go
type InterfaceName interface {
    Method1(param1 Type1) ReturnType1
    Method2(param2 Type2) (ReturnType2, error)
}
```

### 多方法接口

```go
type ReadWriter interface {
    Read(p []byte) (n int, err error)
    Write(p []byte) (n int, err error)
}

type File struct {
    // ...
}

func (f *File) Read(p []byte) (n int, err error) {
    // 实现读取
    return 0, nil
}

func (f *File) Write(p []byte) (n int, err error) {
    // 实现写入
    return 0, nil
}

// File 隐式实现了 ReadWriter 接口
```

---

## 隐式实现

### Go 的隐式实现

Go 不需要显式声明实现了哪个接口，只要类型拥有接口要求的所有方法，就自动实现了该接口。

```go
type Stringer interface {
    String() string
}

type Person struct {
    Name string
    Age  int
}

// Person 实现了 Stringer 接口
// 不需要 "implements Stringer" 声明
func (p Person) String() string {
    return fmt.Sprintf("%s (%d years)", p.Name, p.Age)
}

func main() {
    p := Person{Name: "Alice", Age: 30}

    // Person 可以赋值给 Stringer 接口
    var s Stringer = p
    fmt.Println(s.String()) // Alice (30 years)

    // fmt.Println 会自动调用 String() 方法
    fmt.Println(p) // Alice (30 years)
}
```

### 验证接口实现

编译时验证类型是否实现接口：

```go
// 编译时检查 - 如果 Dog 没有实现 Speaker，编译会失败
var _ Speaker = Dog{}
var _ Speaker = (*Dog)(nil)

// 这是一种常见的惯用法，确保类型实现了接口
```

### 指针接收者与接口

```go
type Modifier interface {
    Modify()
}

type Data struct {
    Value int
}

// 指针接收者方法
func (d *Data) Modify() {
    d.Value = 100
}

func main() {
    d := Data{Value: 1}

    // 错误：Data 没有实现 Modifier，*Data 才实现了
    // var m Modifier = d  // 编译错误

    // 正确：使用指针
    var m Modifier = &d
    m.Modify()
    fmt.Println(d.Value) // 100
}
```

**规则**：
- 值接收者方法：值和指针都可以调用
- 指针接收者方法：只有指针才能满足接口

---

## 空接口

### interface{} 和 any

空接口没有任何方法要求，所以任何类型都实现了空接口。

```go
func main() {
    // interface{} 可以持有任何值
    var anything interface{}

    anything = 42
    fmt.Printf("Type: %T, Value: %v\n", anything, anything)

    anything = "hello"
    fmt.Printf("Type: %T, Value: %v\n", anything, anything)

    anything = []int{1, 2, 3}
    fmt.Printf("Type: %T, Value: %v\n", anything, anything)

    // Go 1.18+ 可以使用 any 作为 interface{} 的别名
    var anyValue any = "Go 1.18+"
    fmt.Println(anyValue)
}
```

### 空接口的使用场景

```go
// 1. 接收任意类型参数
func PrintAnything(v interface{}) {
    fmt.Printf("Value: %v, Type: %T\n", v, v)
}

// 2. 存储任意类型的容器
type Container struct {
    items []interface{}
}

func (c *Container) Add(item interface{}) {
    c.items = append(c.items, item)
}

// 3. JSON 解析未知结构
var data interface{}
json.Unmarshal([]byte(`{"name": "Alice", "age": 30}`), &data)
```

### 空接口的代价

```go
// 使用空接口会失去类型安全
func Sum(a, b interface{}) interface{} {
    // 需要类型断言，运行时可能 panic
    return a.(int) + b.(int)
}

// 更好的方式：使用泛型（Go 1.18+）
func SumGeneric[T int | float64](a, b T) T {
    return a + b
}
```

---

## 接口组合

### 嵌入接口

```go
type Reader interface {
    Read(p []byte) (n int, err error)
}

type Writer interface {
    Write(p []byte) (n int, err error)
}

type Closer interface {
    Close() error
}

// 组合多个接口
type ReadWriter interface {
    Reader
    Writer
}

type ReadWriteCloser interface {
    Reader
    Writer
    Closer
}

// 也可以添加额外方法
type File interface {
    ReadWriteCloser
    Seek(offset int64, whence int) (int64, error)
}
```

### 标准库中的接口组合

```go
// io 包中的接口组合示例
import "io"

// io.ReadWriter 组合了 Reader 和 Writer
type ReadWriter interface {
    Reader
    Writer
}

// io.ReadCloser 组合了 Reader 和 Closer
type ReadCloser interface {
    Reader
    Closer
}
```

---

## 常用标准接口

### fmt.Stringer

```go
type Stringer interface {
    String() string
}

type Point struct {
    X, Y int
}

func (p Point) String() string {
    return fmt.Sprintf("(%d, %d)", p.X, p.Y)
}

func main() {
    p := Point{X: 10, Y: 20}
    fmt.Println(p) // (10, 20)
}
```

### error 接口

```go
type error interface {
    Error() string
}

// 自定义错误类型
type ValidationError struct {
    Field   string
    Message string
}

func (e ValidationError) Error() string {
    return fmt.Sprintf("validation failed on %s: %s", e.Field, e.Message)
}

func Validate(name string) error {
    if name == "" {
        return ValidationError{Field: "name", Message: "cannot be empty"}
    }
    return nil
}
```

### io.Reader 和 io.Writer

```go
import (
    "io"
    "strings"
)

type Reader interface {
    Read(p []byte) (n int, err error)
}

type Writer interface {
    Write(p []byte) (n int, err error)
}

// 使用 io.Reader
func ReadAll(r io.Reader) ([]byte, error) {
    return io.ReadAll(r)
}

func main() {
    // strings.Reader 实现了 io.Reader
    reader := strings.NewReader("Hello, World!")
    data, _ := ReadAll(reader)
    fmt.Println(string(data)) // Hello, World!
}
```

### sort.Interface

```go
import "sort"

type Interface interface {
    Len() int
    Less(i, j int) bool
    Swap(i, j int)
}

type Person struct {
    Name string
    Age  int
}

type ByAge []Person

func (a ByAge) Len() int           { return len(a) }
func (a ByAge) Less(i, j int) bool { return a[i].Age < a[j].Age }
func (a ByAge) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func main() {
    people := []Person{
        {"Alice", 30},
        {"Bob", 25},
        {"Charlie", 35},
    }

    sort.Sort(ByAge(people))
    fmt.Println(people)
    // [{Bob 25} {Alice 30} {Charlie 35}]
}
```

---

## 接口最佳实践

### 1. 接口应该小而专注

```go
// 好 - 小接口
type Reader interface {
    Read(p []byte) (n int, err error)
}

type Writer interface {
    Write(p []byte) (n int, err error)
}

// 避免 - 大而全的接口
type FileOperations interface {
    Read(p []byte) (n int, err error)
    Write(p []byte) (n int, err error)
    Close() error
    Seek(offset int64, whence int) (int64, error)
    Stat() (os.FileInfo, error)
    Sync() error
    // ... 更多方法
}
```

### 2. 在使用方定义接口

```go
// 在使用接口的包中定义，而不是实现方
// 这样可以按需定义最小接口

// 在 userservice 包中
type UserRepository interface {
    GetByID(id int) (*User, error)
    Save(user *User) error
}

type UserService struct {
    repo UserRepository // 依赖接口，而不是具体实现
}

// 任何实现这两个方法的类型都可以作为 repo
```

### 3. 接受接口，返回结构体

```go
// 好 - 参数使用接口
func ProcessData(r io.Reader) error {
    // 可以接受任何 Reader
    return nil
}

// 好 - 返回具体类型
func NewBuffer() *bytes.Buffer {
    return &bytes.Buffer{}
}

// 避免 - 返回接口
func NewReader() io.Reader {
    return &bytes.Buffer{} // 不推荐
}
```

### 4. 避免空接口滥用

```go
// 避免
func DoSomething(data interface{}) {
    // 失去类型安全
}

// 推荐 - 使用具体类型或泛型
func DoSomething(data string) {
    // 类型安全
}

// 或使用泛型 (Go 1.18+)
func DoSomething[T any](data T) {
    // 保持灵活性的同时有类型信息
}
```

---

## 与 JavaScript/TypeScript 对比

### 接口定义对比

```typescript
// TypeScript - 显式接口
interface Speaker {
    speak(): string;
}

class Dog implements Speaker {  // 显式声明实现
    name: string;

    constructor(name: string) {
        this.name = name;
    }

    speak(): string {
        return `${this.name} says: Woof!`;
    }
}
```

```go
// Go - 隐式接口
type Speaker interface {
    Speak() string
}

type Dog struct {
    Name string
}

// 不需要 "implements Speaker"
func (d Dog) Speak() string {
    return d.Name + " says: Woof!"
}
```

### 鸭子类型对比

```typescript
// TypeScript - 结构化类型（鸭子类型）
interface Printable {
    print(): void;
}

function doPrint(p: Printable) {
    p.print();
}

// 任何有 print 方法的对象都可以传入
const obj = {
    print: () => console.log("printing"),
    otherMethod: () => {}
};
doPrint(obj);  // 有效
```

```go
// Go - 类似的鸭子类型
type Printable interface {
    Print()
}

func DoPrint(p Printable) {
    p.Print()
}

type Document struct{}
func (d Document) Print() { fmt.Println("printing") }

// Document 自动满足 Printable 接口
doc := Document{}
DoPrint(doc)
```

### 空接口 vs any

```typescript
// TypeScript
function process(data: any): void {
    // 任何类型
}

function processUnknown(data: unknown): void {
    // 更安全，需要类型检查才能使用
    if (typeof data === "string") {
        console.log(data.toUpperCase());
    }
}
```

```go
// Go
func Process(data interface{}) {
    // 任何类型
}

func ProcessAny(data any) {
    // Go 1.18+，any 是 interface{} 的别名
    // 需要类型断言才能使用具体方法
    if s, ok := data.(string); ok {
        fmt.Println(strings.ToUpper(s))
    }
}
```

### 主要差异

| 特性 | TypeScript | Go |
|-----|-----------|-----|
| 实现声明 | `implements` (显式) | 隐式 |
| 接口检查 | 编译时 | 编译时 |
| 空接口 | `any`, `unknown` | `interface{}`, `any` |
| 接口组合 | `extends` | 嵌入 |
| 可选方法 | 支持 (`method?()`) | 不支持 |
| 默认实现 | 不支持 | 不支持 |

---

## 下一步

- [类型断言](./05-type-assertion.md) - 学习类型断言和类型开关
