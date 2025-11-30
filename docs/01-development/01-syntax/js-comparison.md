# Go vs JavaScript/TypeScript 对比总结

本文档汇总 Go 与 JavaScript/TypeScript 的主要差异，帮助前端开发者快速掌握 Go。

## 目录

- [语言特性对比](#语言特性对比)
- [类型系统](#类型系统)
- [变量与常量](#变量与常量)
- [函数](#函数)
- [控制流](#控制流)
- [数据结构](#数据结构)
- [错误处理](#错误处理)
- [并发模型](#并发模型)
- [包管理](#包管理)
- [常见陷阱](#常见陷阱)
- [速查表](#速查表)

---

## 语言特性对比

| 特性 | JavaScript/TypeScript | Go |
|-----|----------------------|-----|
| 类型系统 | 动态(JS) / 静态可选(TS) | 静态强类型 |
| 编译方式 | JIT / AOT(TS) | AOT 编译为原生二进制 |
| 运行时 | Node.js / 浏览器 | 无需运行时 |
| 垃圾回收 | 有 | 有 |
| 并发模型 | 事件循环 + Promise | goroutine + channel |
| 面向对象 | 原型/类继承 | 组合 + 接口 |
| 包管理 | npm/yarn | Go Modules |
| 空值 | null, undefined | nil |
| 泛型 | 支持(TS) | Go 1.18+ 支持 |

---

## 类型系统

### 基本类型

| JavaScript | TypeScript | Go |
|-----------|-----------|-----|
| `number` | `number` | `int`, `int64`, `float64` |
| `string` | `string` | `string` |
| `boolean` | `boolean` | `bool` |
| `null` | `null` | `nil` (仅用于引用类型) |
| `undefined` | `undefined` | 无 (使用零值) |
| `bigint` | `bigint` | `int64`, `big.Int` |
| `symbol` | `symbol` | 无 |

### 类型声明

```typescript
// TypeScript
let name: string = "Gopher";
let age: number = 25;
let isActive: boolean = true;
let data: any = "anything";

// 类型推断
let inferred = "hello";  // string
```

```go
// Go
var name string = "Gopher"
var age int = 25
var isActive bool = true
var data interface{} = "anything"  // 或 any (Go 1.18+)

// 类型推断
inferred := "hello"  // string
```

### 类型转换

```typescript
// TypeScript - 类型断言
const value: any = "hello";
const str = value as string;
const str2 = <string>value;

// 运行时转换
const num = Number("42");
const str3 = String(42);
```

```go
// Go - 类型转换/断言
var i int = 42
var f float64 = float64(i)  // 类型转换

var value interface{} = "hello"
str := value.(string)              // 类型断言（可能 panic）
str, ok := value.(string)          // 安全断言
```

---

## 变量与常量

### 变量声明

```typescript
// TypeScript
let name = "Gopher";       // 可变
const PI = 3.14;           // 不可变引用
var legacy = "old";        // 函数作用域

// 解构
const { a, b } = obj;
const [x, y] = arr;
```

```go
// Go
var name string = "Gopher"  // 显式类型
name := "Gopher"            // 短声明（函数内）
const PI = 3.14             // 编译时常量

// 无解构语法，但多返回值可以
a, b := getValues()
```

### 零值 vs undefined

```typescript
// TypeScript
let value: number;
console.log(value);  // undefined (严格模式会报错)

let obj: { name?: string } = {};
console.log(obj.name);  // undefined
```

```go
// Go - 零值
var i int       // 0
var f float64   // 0.0
var s string    // ""
var b bool      // false
var p *int      // nil

m := map[string]int{}
fmt.Println(m["key"])  // 0 (零值，不是 undefined)
```

---

## 函数

### 函数定义

```typescript
// TypeScript
function add(a: number, b: number): number {
    return a + b;
}

// 箭头函数
const multiply = (a: number, b: number): number => a * b;

// 可选参数
function greet(name: string, greeting?: string): string {
    return `${greeting || "Hello"}, ${name}!`;
}

// 默认参数
function greet2(name: string, greeting = "Hello"): string {
    return `${greeting}, ${name}!`;
}
```

```go
// Go
func add(a, b int) int {
    return a + b
}

// 匿名函数
multiply := func(a, b int) int {
    return a * b
}

// 无可选参数/默认参数，使用变通方法
func greet(name string, opts ...string) string {
    greeting := "Hello"
    if len(opts) > 0 {
        greeting = opts[0]
    }
    return greeting + ", " + name + "!"
}
```

### 多返回值

```typescript
// TypeScript - 使用对象或元组
function divide(a: number, b: number): { result: number; error: string | null } {
    if (b === 0) return { result: 0, error: "division by zero" };
    return { result: a / b, error: null };
}

// 解构
const { result, error } = divide(10, 2);
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

---

## 控制流

### 条件语句

```typescript
// TypeScript
if (value) {          // truthy/falsy
    // ...
}

// 三元运算符
const result = condition ? "yes" : "no";
```

```go
// Go
if value != 0 {       // 必须是 bool
    // ...
}

// 无三元运算符
var result string
if condition {
    result = "yes"
} else {
    result = "no"
}
```

### 循环

```typescript
// TypeScript
for (let i = 0; i < 10; i++) {}
for (const item of array) {}
for (const key in object) {}
while (condition) {}
array.forEach(item => {});
```

```go
// Go - 只有 for
for i := 0; i < 10; i++ {}
for _, item := range slice {}
for key, value := range map {}
for condition {}  // 类似 while
for {}            // 无限循环
```

### switch

```typescript
// TypeScript
switch (value) {
    case 1:
        console.log("one");
        break;  // 必须 break
    case 2:
    case 3:
        console.log("two or three");
        break;
    default:
        console.log("other");
}
```

```go
// Go
switch value {
case 1:
    fmt.Println("one")
    // 自动 break
case 2, 3:
    fmt.Println("two or three")
default:
    fmt.Println("other")
}
```

---

## 数据结构

### 数组和切片

```typescript
// TypeScript
const arr: number[] = [1, 2, 3];
arr.push(4);
arr.pop();
const filtered = arr.filter(x => x > 1);
const mapped = arr.map(x => x * 2);
```

```go
// Go
arr := [3]int{1, 2, 3}     // 数组 (固定长度)
slice := []int{1, 2, 3}     // 切片 (动态)
slice = append(slice, 4)
// 无内置 filter/map，需要手动实现
var filtered []int
for _, x := range slice {
    if x > 1 {
        filtered = append(filtered, x)
    }
}
```

### 对象和映射

```typescript
// TypeScript
interface User {
    name: string;
    age: number;
}

const user: User = { name: "Alice", age: 30 };
const map = new Map<string, number>();
map.set("key", 1);
map.get("key");
map.has("key");
```

```go
// Go
type User struct {
    Name string
    Age  int
}

user := User{Name: "Alice", Age: 30}
m := map[string]int{}
m["key"] = 1
value := m["key"]
value, exists := m["key"]
```

### 类和结构体

```typescript
// TypeScript
class User {
    private name: string;

    constructor(name: string) {
        this.name = name;
    }

    greet(): string {
        return `Hello, ${this.name}!`;
    }
}

const user = new User("Alice");
```

```go
// Go
type User struct {
    name string  // 小写 = 私有
}

func NewUser(name string) *User {
    return &User{name: name}
}

func (u *User) Greet() string {
    return "Hello, " + u.name + "!"
}

user := NewUser("Alice")
```

---

## 错误处理

```typescript
// TypeScript
async function fetchData(): Promise<Data> {
    try {
        const response = await fetch(url);
        if (!response.ok) {
            throw new Error("Failed to fetch");
        }
        return await response.json();
    } catch (error) {
        console.error("Error:", error);
        throw error;
    }
}
```

```go
// Go
func fetchData() (Data, error) {
    resp, err := http.Get(url)
    if err != nil {
        return Data{}, fmt.Errorf("fetch failed: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != 200 {
        return Data{}, errors.New("bad status")
    }

    var data Data
    if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
        return Data{}, fmt.Errorf("decode failed: %w", err)
    }
    return data, nil
}
```

---

## 并发模型

```typescript
// TypeScript - Promise/async-await
async function fetchAll(urls: string[]): Promise<Response[]> {
    return Promise.all(urls.map(url => fetch(url)));
}

// 执行
const results = await fetchAll(["url1", "url2"]);
```

```go
// Go - goroutine + channel
func fetchAll(urls []string) []Response {
    ch := make(chan Response, len(urls))

    for _, url := range urls {
        go func(u string) {
            resp := fetch(u)
            ch <- resp
        }(url)
    }

    results := make([]Response, len(urls))
    for i := range urls {
        results[i] = <-ch
    }
    return results
}
```

---

## 包管理

### 导入导出

```typescript
// TypeScript
// utils.ts
export function add(a: number, b: number): number {
    return a + b;
}
export const PI = 3.14;
export default class Calculator {}

// main.ts
import Calculator, { add, PI } from './utils';
```

```go
// Go
// utils/utils.go
package utils

func Add(a, b int) int {  // 大写 = 导出
    return a + b
}
const PI = 3.14           // 大写 = 导出

// main.go
package main
import "myproject/utils"

func main() {
    sum := utils.Add(1, 2)
}
```

### 包管理命令

| npm/yarn | Go |
|----------|-----|
| `npm init` | `go mod init` |
| `npm install pkg` | `go get pkg` |
| `npm install` | `go mod download` |
| `npm update` | `go get -u` |
| `npm prune` | `go mod tidy` |
| `npm run build` | `go build` |
| `npm test` | `go test ./...` |

---

## 常见陷阱

### 1. 未使用的变量

```go
// 编译错误！
func main() {
    x := 10  // declared but not used
}

// 解决方案
_ = x  // 使用空白标识符
```

### 2. 短声明变量作用域

```go
// 陷阱
var x int = 10
if true {
    x := 20  // 这是新变量！
}
fmt.Println(x)  // 仍然是 10

// 解决方案
if true {
    x = 20  // 使用 = 而不是 :=
}
```

### 3. 循环变量捕获

```go
// 陷阱
for _, v := range values {
    go func() {
        fmt.Println(v)  // 可能全部打印最后一个值
    }()
}

// 解决方案
for _, v := range values {
    v := v  // 创建新变量
    go func() {
        fmt.Println(v)
    }()
}
```

### 4. nil map 写入

```go
// panic!
var m map[string]int
m["key"] = 1  // panic: assignment to nil map

// 解决方案
m := make(map[string]int)
m["key"] = 1
```

### 5. 字符串长度

```go
s := "你好"
len(s)           // 6 (字节数)
len([]rune(s))   // 2 (字符数)
```

---

## 速查表

### 类型对照

| TypeScript | Go |
|-----------|-----|
| `number` | `int`, `float64` |
| `string` | `string` |
| `boolean` | `bool` |
| `any` | `interface{}` 或 `any` |
| `T[]` | `[]T` |
| `Record<K, V>` | `map[K]V` |
| `Partial<T>` | 无直接对应 |
| `Promise<T>` | `chan T`, 或返回值 |
| `null \| undefined` | `nil` |

### 常用操作对照

| 操作 | TypeScript | Go |
|-----|-----------|-----|
| 打印 | `console.log()` | `fmt.Println()` |
| 格式化字符串 | `` `${var}` `` | `fmt.Sprintf()` |
| 数组添加 | `arr.push(x)` | `arr = append(arr, x)` |
| 数组长度 | `arr.length` | `len(arr)` |
| 字典检查 | `key in obj` | `_, ok := m[key]` |
| 类型检查 | `typeof x` | `reflect.TypeOf(x)` |
| 异步 | `async/await` | goroutine |
| 捕获错误 | `try/catch` | `if err != nil` |

---

## 推荐学习路径

1. **第一周**: 语法基础 (变量、类型、函数、控制流)
2. **第二周**: 复合类型 (结构体、切片、映射、接口)
3. **第三周**: 并发编程 (goroutine、channel、sync)
4. **第四周**: 标准库 (net/http、encoding/json、io)
5. **后续**: 框架学习 (Gin、GORM 等)

---

## 参考资源

- [Go 语言之旅](https://tour.go-zh.org/)
- [Go by Example](https://gobyexample.com/)
- [Effective Go](https://go.dev/doc/effective_go)
- [Go 标准库文档](https://pkg.go.dev/std)
