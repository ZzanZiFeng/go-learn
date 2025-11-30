# 结构体与接口练习

## 练习 1：定义结构体

创建一个 `Book` 结构体，包含以下字段：
- ISBN (string)
- Title (string)
- Author (string)
- Price (float64)
- PublishedYear (int)

要求：
1. 使用适当的 JSON 标签
2. 创建工厂函数 `NewBook`
3. 实现 `String()` 方法

<details>
<summary>查看答案</summary>

```go
package main

import (
	"fmt"
)

type Book struct {
	ISBN          string  `json:"isbn"`
	Title         string  `json:"title"`
	Author        string  `json:"author"`
	Price         float64 `json:"price"`
	PublishedYear int     `json:"published_year"`
}

func NewBook(isbn, title, author string, price float64, year int) *Book {
	return &Book{
		ISBN:          isbn,
		Title:         title,
		Author:        author,
		Price:         price,
		PublishedYear: year,
	}
}

func (b Book) String() string {
	return fmt.Sprintf("%s by %s (%d) - $%.2f", b.Title, b.Author, b.PublishedYear, b.Price)
}

func main() {
	book := NewBook("978-0-13-468599-1", "The Go Programming Language", "Alan Donovan", 49.99, 2015)
	fmt.Println(book)
	// Output: The Go Programming Language by Alan Donovan (2015) - $49.99
}
```

</details>

---

## 练习 2：值接收者 vs 指针接收者

创建一个 `BankAccount` 结构体，实现以下方法：
- `Balance()` - 返回余额（值接收者）
- `Deposit(amount float64)` - 存款（指针接收者）
- `Withdraw(amount float64) error` - 取款（指针接收者，余额不足返回错误）

<details>
<summary>查看答案</summary>

```go
package main

import (
	"errors"
	"fmt"
)

type BankAccount struct {
	owner   string
	balance float64
}

func NewBankAccount(owner string, initial float64) *BankAccount {
	return &BankAccount{
		owner:   owner,
		balance: initial,
	}
}

// 值接收者 - 只读操作
func (a BankAccount) Balance() float64 {
	return a.balance
}

func (a BankAccount) Owner() string {
	return a.owner
}

// 指针接收者 - 修改操作
func (a *BankAccount) Deposit(amount float64) {
	if amount > 0 {
		a.balance += amount
	}
}

func (a *BankAccount) Withdraw(amount float64) error {
	if amount <= 0 {
		return errors.New("amount must be positive")
	}
	if amount > a.balance {
		return errors.New("insufficient funds")
	}
	a.balance -= amount
	return nil
}

func main() {
	account := NewBankAccount("Alice", 1000)

	fmt.Printf("Owner: %s, Balance: $%.2f\n", account.Owner(), account.Balance())
	// Output: Owner: Alice, Balance: $1000.00

	account.Deposit(500)
	fmt.Printf("After deposit: $%.2f\n", account.Balance())
	// Output: After deposit: $1500.00

	err := account.Withdraw(2000)
	if err != nil {
		fmt.Println("Withdraw error:", err)
	}
	// Output: Withdraw error: insufficient funds

	account.Withdraw(300)
	fmt.Printf("After withdraw: $%.2f\n", account.Balance())
	// Output: After withdraw: $1200.00
}
```

</details>

---

## 练习 3：接口实现

定义一个 `Notifier` 接口，包含 `Notify(message string) error` 方法。

实现三种通知器：
- `EmailNotifier` - 打印 "Sending email: <message>"
- `SMSNotifier` - 打印 "Sending SMS: <message>"
- `PushNotifier` - 打印 "Sending push: <message>"

创建一个函数 `SendAll(notifiers []Notifier, message string)` 发送通知。

<details>
<summary>查看答案</summary>

```go
package main

import "fmt"

type Notifier interface {
	Notify(message string) error
}

type EmailNotifier struct {
	Email string
}

func (e EmailNotifier) Notify(message string) error {
	fmt.Printf("Sending email to %s: %s\n", e.Email, message)
	return nil
}

type SMSNotifier struct {
	Phone string
}

func (s SMSNotifier) Notify(message string) error {
	fmt.Printf("Sending SMS to %s: %s\n", s.Phone, message)
	return nil
}

type PushNotifier struct {
	DeviceID string
}

func (p PushNotifier) Notify(message string) error {
	fmt.Printf("Sending push to device %s: %s\n", p.DeviceID, message)
	return nil
}

func SendAll(notifiers []Notifier, message string) {
	for _, n := range notifiers {
		n.Notify(message)
	}
}

func main() {
	notifiers := []Notifier{
		EmailNotifier{Email: "user@example.com"},
		SMSNotifier{Phone: "+1234567890"},
		PushNotifier{DeviceID: "device-123"},
	}

	SendAll(notifiers, "Hello, World!")
	// Output:
	// Sending email to user@example.com: Hello, World!
	// Sending SMS to +1234567890: Hello, World!
	// Sending push to device device-123: Hello, World!
}
```

</details>

---

## 练习 4：类型断言和类型开关

编写一个 `Describe` 函数，接受 `interface{}` 参数，根据类型打印不同信息：
- `int`: "Integer: X, squared: X*X"
- `string`: "String: 'X', length: N"
- `bool`: "Boolean: X"
- `[]int`: "Int slice with N elements"
- 其他: "Unknown type: T"

<details>
<summary>查看答案</summary>

```go
package main

import "fmt"

func Describe(v interface{}) {
	switch val := v.(type) {
	case int:
		fmt.Printf("Integer: %d, squared: %d\n", val, val*val)
	case string:
		fmt.Printf("String: '%s', length: %d\n", val, len(val))
	case bool:
		fmt.Printf("Boolean: %t\n", val)
	case []int:
		fmt.Printf("Int slice with %d elements: %v\n", len(val), val)
	default:
		fmt.Printf("Unknown type: %T\n", val)
	}
}

func main() {
	Describe(42)
	Describe("hello")
	Describe(true)
	Describe([]int{1, 2, 3, 4, 5})
	Describe(3.14)
	// Output:
	// Integer: 42, squared: 1764
	// String: 'hello', length: 5
	// Boolean: true
	// Int slice with 5 elements: [1 2 3 4 5]
	// Unknown type: float64
}
```

</details>

---

## 练习 5：结构体嵌入

创建以下类型层次：
1. `Timestamp` 结构体：CreatedAt, UpdatedAt (time.Time)
2. `Entity` 结构体：ID (int), 嵌入 Timestamp
3. `User` 结构体：嵌入 Entity, Name, Email

实现：
- `Touch()` 方法更新 UpdatedAt
- `User` 的 `String()` 方法

<details>
<summary>查看答案</summary>

```go
package main

import (
	"fmt"
	"time"
)

type Timestamp struct {
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (t *Timestamp) Touch() {
	t.UpdatedAt = time.Now()
}

type Entity struct {
	ID int
	Timestamp
}

type User struct {
	Entity
	Name  string
	Email string
}

func (u User) String() string {
	return fmt.Sprintf("User #%d: %s <%s>", u.ID, u.Name, u.Email)
}

func NewUser(id int, name, email string) *User {
	now := time.Now()
	return &User{
		Entity: Entity{
			ID: id,
			Timestamp: Timestamp{
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
		Name:  name,
		Email: email,
	}
}

func main() {
	user := NewUser(1, "Alice", "alice@example.com")
	fmt.Println(user)
	// Output: User #1: Alice <alice@example.com>

	// Access promoted fields
	fmt.Println("ID:", user.ID)
	fmt.Println("Created:", user.CreatedAt.Format("2006-01-02 15:04:05"))

	// Use promoted method
	time.Sleep(10 * time.Millisecond)
	user.Touch()
	fmt.Println("Updated:", user.UpdatedAt.Format("2006-01-02 15:04:05.000"))
}
```

</details>

---

## 练习 6：泛型函数

实现以下泛型函数：
1. `Map[T, U any](slice []T, f func(T) U) []U` - 映射切片
2. `Filter[T any](slice []T, predicate func(T) bool) []T` - 过滤切片
3. `Reduce[T, U any](slice []T, initial U, f func(U, T) U) U` - 归约切片

<details>
<summary>查看答案</summary>

```go
package main

import "fmt"

func Map[T, U any](slice []T, f func(T) U) []U {
	result := make([]U, len(slice))
	for i, v := range slice {
		result[i] = f(v)
	}
	return result
}

func Filter[T any](slice []T, predicate func(T) bool) []T {
	result := make([]T, 0)
	for _, v := range slice {
		if predicate(v) {
			result = append(result, v)
		}
	}
	return result
}

func Reduce[T, U any](slice []T, initial U, f func(U, T) U) U {
	result := initial
	for _, v := range slice {
		result = f(result, v)
	}
	return result
}

func main() {
	nums := []int{1, 2, 3, 4, 5}

	// Map: double each number
	doubled := Map(nums, func(n int) int { return n * 2 })
	fmt.Println("Doubled:", doubled)
	// Output: Doubled: [2 4 6 8 10]

	// Map: int to string
	strings := Map(nums, func(n int) string { return fmt.Sprintf("#%d", n) })
	fmt.Println("Strings:", strings)
	// Output: Strings: [#1 #2 #3 #4 #5]

	// Filter: keep even numbers
	evens := Filter(nums, func(n int) bool { return n%2 == 0 })
	fmt.Println("Evens:", evens)
	// Output: Evens: [2 4]

	// Filter: keep numbers > 3
	big := Filter(nums, func(n int) bool { return n > 3 })
	fmt.Println("Greater than 3:", big)
	// Output: Greater than 3: [4 5]

	// Reduce: sum
	sum := Reduce(nums, 0, func(acc, n int) int { return acc + n })
	fmt.Println("Sum:", sum)
	// Output: Sum: 15

	// Reduce: product
	product := Reduce(nums, 1, func(acc, n int) int { return acc * n })
	fmt.Println("Product:", product)
	// Output: Product: 120

	// Combine: sum of doubled evens
	result := Reduce(
		Map(
			Filter(nums, func(n int) bool { return n%2 == 0 }),
			func(n int) int { return n * 2 },
		),
		0,
		func(acc, n int) int { return acc + n },
	)
	fmt.Println("Sum of doubled evens:", result)
	// Output: Sum of doubled evens: 12 (2*2 + 4*2 = 4 + 8)
}
```

</details>

---

## 练习 7：泛型类型

实现一个泛型 `Stack[T any]` 类型，支持以下操作：
- `Push(item T)`
- `Pop() (T, bool)`
- `Peek() (T, bool)`
- `Len() int`
- `IsEmpty() bool`

<details>
<summary>查看答案</summary>

```go
package main

import "fmt"

type Stack[T any] struct {
	items []T
}

func NewStack[T any]() *Stack[T] {
	return &Stack[T]{
		items: make([]T, 0),
	}
}

func (s *Stack[T]) Push(item T) {
	s.items = append(s.items, item)
}

func (s *Stack[T]) Pop() (T, bool) {
	if s.IsEmpty() {
		var zero T
		return zero, false
	}
	idx := len(s.items) - 1
	item := s.items[idx]
	s.items = s.items[:idx]
	return item, true
}

func (s *Stack[T]) Peek() (T, bool) {
	if s.IsEmpty() {
		var zero T
		return zero, false
	}
	return s.items[len(s.items)-1], true
}

func (s *Stack[T]) Len() int {
	return len(s.items)
}

func (s *Stack[T]) IsEmpty() bool {
	return len(s.items) == 0
}

func main() {
	// Integer stack
	intStack := NewStack[int]()
	intStack.Push(1)
	intStack.Push(2)
	intStack.Push(3)

	fmt.Println("Stack length:", intStack.Len())
	// Output: Stack length: 3

	if val, ok := intStack.Peek(); ok {
		fmt.Println("Peek:", val)
	}
	// Output: Peek: 3

	for !intStack.IsEmpty() {
		if val, ok := intStack.Pop(); ok {
			fmt.Println("Popped:", val)
		}
	}
	// Output:
	// Popped: 3
	// Popped: 2
	// Popped: 1

	// String stack
	strStack := NewStack[string]()
	strStack.Push("hello")
	strStack.Push("world")

	val, _ := strStack.Pop()
	fmt.Println("String popped:", val)
	// Output: String popped: world
}
```

</details>

---

## 练习 8：实现 sort.Interface

创建一个 `Product` 结构体（Name, Price, Rating），实现 `sort.Interface` 以支持按不同字段排序。

<details>
<summary>查看答案</summary>

```go
package main

import (
	"fmt"
	"sort"
)

type Product struct {
	Name   string
	Price  float64
	Rating float64
}

// ByPrice sorts products by price (ascending)
type ByPrice []Product

func (p ByPrice) Len() int           { return len(p) }
func (p ByPrice) Less(i, j int) bool { return p[i].Price < p[j].Price }
func (p ByPrice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// ByRating sorts products by rating (descending - highest first)
type ByRating []Product

func (p ByRating) Len() int           { return len(p) }
func (p ByRating) Less(i, j int) bool { return p[i].Rating > p[j].Rating }
func (p ByRating) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// ByName sorts products by name (alphabetically)
type ByName []Product

func (p ByName) Len() int           { return len(p) }
func (p ByName) Less(i, j int) bool { return p[i].Name < p[j].Name }
func (p ByName) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func printProducts(label string, products []Product) {
	fmt.Println(label)
	for _, p := range products {
		fmt.Printf("  %s: $%.2f (%.1f★)\n", p.Name, p.Price, p.Rating)
	}
}

func main() {
	products := []Product{
		{Name: "Laptop", Price: 999.99, Rating: 4.5},
		{Name: "Mouse", Price: 29.99, Rating: 4.8},
		{Name: "Keyboard", Price: 79.99, Rating: 4.2},
		{Name: "Monitor", Price: 299.99, Rating: 4.6},
	}

	printProducts("Original:", products)

	// Sort by price
	sort.Sort(ByPrice(products))
	printProducts("\nBy Price (ascending):", products)

	// Sort by rating
	sort.Sort(ByRating(products))
	printProducts("\nBy Rating (descending):", products)

	// Sort by name
	sort.Sort(ByName(products))
	printProducts("\nBy Name:", products)
}
```

</details>

---

## 挑战练习：简单 ORM

创建一个简化的 ORM 风格的查询构建器：

```go
type QueryBuilder[T any] struct { ... }

// 支持链式调用
query := NewQuery[User]().
    Where("age", ">", 18).
    Where("status", "=", "active").
    OrderBy("name", "ASC").
    Limit(10)

sql := query.ToSQL()
// SELECT * FROM users WHERE age > 18 AND status = 'active' ORDER BY name ASC LIMIT 10
```

<details>
<summary>查看答案</summary>

```go
package main

import (
	"fmt"
	"strings"
)

type Condition struct {
	Field    string
	Operator string
	Value    interface{}
}

type Order struct {
	Field     string
	Direction string
}

type QueryBuilder[T any] struct {
	tableName  string
	conditions []Condition
	orders     []Order
	limit      int
	offset     int
}

func NewQuery[T any](table string) *QueryBuilder[T] {
	return &QueryBuilder[T]{
		tableName:  table,
		conditions: make([]Condition, 0),
		orders:     make([]Order, 0),
	}
}

func (q *QueryBuilder[T]) Where(field, operator string, value interface{}) *QueryBuilder[T] {
	q.conditions = append(q.conditions, Condition{
		Field:    field,
		Operator: operator,
		Value:    value,
	})
	return q
}

func (q *QueryBuilder[T]) OrderBy(field, direction string) *QueryBuilder[T] {
	q.orders = append(q.orders, Order{
		Field:     field,
		Direction: direction,
	})
	return q
}

func (q *QueryBuilder[T]) Limit(n int) *QueryBuilder[T] {
	q.limit = n
	return q
}

func (q *QueryBuilder[T]) Offset(n int) *QueryBuilder[T] {
	q.offset = n
	return q
}

func (q *QueryBuilder[T]) ToSQL() string {
	var parts []string

	// SELECT
	parts = append(parts, fmt.Sprintf("SELECT * FROM %s", q.tableName))

	// WHERE
	if len(q.conditions) > 0 {
		var whereParts []string
		for _, c := range q.conditions {
			var valueStr string
			switch v := c.Value.(type) {
			case string:
				valueStr = fmt.Sprintf("'%s'", v)
			default:
				valueStr = fmt.Sprintf("%v", v)
			}
			whereParts = append(whereParts, fmt.Sprintf("%s %s %s", c.Field, c.Operator, valueStr))
		}
		parts = append(parts, "WHERE "+strings.Join(whereParts, " AND "))
	}

	// ORDER BY
	if len(q.orders) > 0 {
		var orderParts []string
		for _, o := range q.orders {
			orderParts = append(orderParts, fmt.Sprintf("%s %s", o.Field, o.Direction))
		}
		parts = append(parts, "ORDER BY "+strings.Join(orderParts, ", "))
	}

	// LIMIT
	if q.limit > 0 {
		parts = append(parts, fmt.Sprintf("LIMIT %d", q.limit))
	}

	// OFFSET
	if q.offset > 0 {
		parts = append(parts, fmt.Sprintf("OFFSET %d", q.offset))
	}

	return strings.Join(parts, " ")
}

type User struct {
	ID     int
	Name   string
	Age    int
	Status string
}

func main() {
	query := NewQuery[User]("users").
		Where("age", ">", 18).
		Where("status", "=", "active").
		OrderBy("name", "ASC").
		OrderBy("created_at", "DESC").
		Limit(10).
		Offset(20)

	fmt.Println(query.ToSQL())
	// Output: SELECT * FROM users WHERE age > 18 AND status = 'active' ORDER BY name ASC, created_at DESC LIMIT 10 OFFSET 20

	// Simple query
	simple := NewQuery[User]("users").
		Where("id", "=", 1).
		Limit(1)

	fmt.Println(simple.ToSQL())
	// Output: SELECT * FROM users WHERE id = 1 LIMIT 1
}
```

</details>

---

## 下一步

完成这些练习后，你应该对 Go 的结构体和接口有了扎实的理解。继续学习：

- [并发编程](../03-concurrency/) - 学习 goroutine 和 channel
