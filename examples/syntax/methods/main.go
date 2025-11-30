// Package main demonstrates methods in Go
package main

import (
	"fmt"
	"math"
)

// Counter demonstrates value vs pointer receivers
type Counter struct {
	Value int
}

// Increment uses a value receiver - doesn't modify the original
func (c Counter) Increment() {
	c.Value++ // Only modifies the copy
}

// IncrementPtr uses a pointer receiver - modifies the original
func (c *Counter) IncrementPtr() {
	c.Value++ // Modifies the actual Counter
}

// GetValue is a value receiver for read-only operations
func (c Counter) GetValue() int {
	return c.Value
}

// Rectangle demonstrates multiple methods on a type
type Rectangle struct {
	Width  float64
	Height float64
}

// Area calculates the area (value receiver - read-only)
func (r Rectangle) Area() float64 {
	return r.Width * r.Height
}

// Perimeter calculates the perimeter (value receiver - read-only)
func (r Rectangle) Perimeter() float64 {
	return 2 * (r.Width + r.Height)
}

// Scale modifies the dimensions (pointer receiver - modifies)
func (r *Rectangle) Scale(factor float64) {
	r.Width *= factor
	r.Height *= factor
}

// Circle demonstrates methods for a different shape
type Circle struct {
	Radius float64
}

// Area calculates the circle area
func (c Circle) Area() float64 {
	return math.Pi * c.Radius * c.Radius
}

// User demonstrates real-world method usage
type User struct {
	Name  string
	Email string
	Age   int
}

// String returns a string representation (implements fmt.Stringer)
func (u User) String() string {
	return fmt.Sprintf("%s <%s>", u.Name, u.Email)
}

// IsAdult checks if the user is an adult
func (u User) IsAdult() bool {
	return u.Age >= 18
}

// SetName modifies the user's name
func (u *User) SetName(name string) {
	u.Name = name
}

// HaveBirthday increments the user's age
func (u *User) HaveBirthday() {
	u.Age++
}

// MyInt demonstrates methods on custom types
type MyInt int

// Double returns twice the value
func (m MyInt) Double() MyInt {
	return m * 2
}

// IsPositive checks if the value is positive
func (m MyInt) IsPositive() bool {
	return m > 0
}

// Builder demonstrates method chaining
type Builder struct {
	value string
}

// Append adds text and returns the builder for chaining
func (b *Builder) Append(s string) *Builder {
	b.value += s
	return b
}

// AppendLine adds text with newline
func (b *Builder) AppendLine(s string) *Builder {
	b.value += s + "\n"
	return b
}

// Build returns the final string
func (b *Builder) Build() string {
	return b.value
}

// Clear resets the builder
func (b *Builder) Clear() *Builder {
	b.value = ""
	return b
}

// Adder demonstrates method values and expressions
type Adder struct {
	Base int
}

// Add adds a number to the base
func (a Adder) Add(n int) int {
	return a.Base + n
}

func main() {
	fmt.Println("=== Methods in Go ===")

	// Value receiver vs Pointer receiver
	fmt.Println("\n--- Value vs Pointer Receivers ---")

	counter := Counter{Value: 0}

	// Value receiver doesn't modify original
	counter.Increment()
	fmt.Println("After Increment():", counter.Value)
	// Output: After Increment(): 0

	// Pointer receiver modifies original
	counter.IncrementPtr()
	fmt.Println("After IncrementPtr():", counter.Value)
	// Output: After IncrementPtr(): 1

	// Go automatically converts between value and pointer
	fmt.Println("\n--- Automatic Conversion ---")
	counterPtr := &Counter{Value: 10}

	// Pointer can call value receiver methods
	fmt.Println("Pointer calling GetValue():", counterPtr.GetValue())
	// Output: Pointer calling GetValue(): 10

	// Value can call pointer receiver methods (Go auto-converts)
	counter2 := Counter{Value: 5}
	counter2.IncrementPtr() // Automatically becomes (&counter2).IncrementPtr()
	fmt.Println("Value calling IncrementPtr():", counter2.Value)
	// Output: Value calling IncrementPtr(): 6

	// Rectangle methods
	fmt.Println("\n--- Rectangle Methods ---")
	rect := Rectangle{Width: 10, Height: 5}
	fmt.Println("Area:", rect.Area())
	fmt.Println("Perimeter:", rect.Perimeter())
	// Output:
	// Area: 50
	// Perimeter: 30

	rect.Scale(2)
	fmt.Println("After Scale(2):")
	fmt.Println("  Width:", rect.Width)
	fmt.Println("  Height:", rect.Height)
	fmt.Println("  New Area:", rect.Area())
	// Output:
	// After Scale(2):
	//   Width: 20
	//   Height: 10
	//   New Area: 200

	// Same method name, different types
	fmt.Println("\n--- Same Method Name, Different Types ---")
	circle := Circle{Radius: 5}
	fmt.Printf("Circle Area: %.2f\n", circle.Area())
	fmt.Printf("Rectangle Area: %.2f\n", rect.Area())
	// Output:
	// Circle Area: 78.54
	// Rectangle Area: 200.00

	// User methods
	fmt.Println("\n--- User Methods ---")
	user := User{Name: "Alice", Email: "alice@example.com", Age: 17}

	// String() is called automatically by fmt
	fmt.Println("User:", user)
	// Output: User: Alice <alice@example.com>

	fmt.Println("Is Adult:", user.IsAdult())
	// Output: Is Adult: false

	user.HaveBirthday()
	fmt.Println("After birthday, Is Adult:", user.IsAdult())
	// Output: After birthday, Is Adult: true

	user.SetName("Alice Smith")
	fmt.Println("After SetName:", user)
	// Output: After SetName: Alice Smith <alice@example.com>

	// Methods on custom types
	fmt.Println("\n--- Methods on Custom Types ---")
	num := MyInt(5)
	fmt.Println("MyInt:", num)
	fmt.Println("Double:", num.Double())
	fmt.Println("IsPositive:", num.IsPositive())
	// Output:
	// MyInt: 5
	// Double: 10
	// IsPositive: true

	negNum := MyInt(-3)
	fmt.Println("Negative IsPositive:", negNum.IsPositive())
	// Output: Negative IsPositive: false

	// Method chaining
	fmt.Println("\n--- Method Chaining ---")
	builder := &Builder{}
	result := builder.
		Append("Hello").
		Append(" ").
		Append("World").
		AppendLine("!").
		Append("Go is awesome").
		Build()

	fmt.Println("Builder result:")
	fmt.Println(result)
	// Output:
	// Builder result:
	// Hello World!
	// Go is awesome

	// Method values
	fmt.Println("\n--- Method Values ---")
	adder := Adder{Base: 10}

	// Method value - bound to specific receiver
	addFunc := adder.Add
	fmt.Println("Method value:", addFunc(5))  // 15
	fmt.Println("Method value:", addFunc(10)) // 20
	// Output:
	// Method value: 15
	// Method value: 20

	// Method expression
	fmt.Println("\n--- Method Expressions ---")
	addExpr := Adder.Add // Method expression - needs explicit receiver

	adder1 := Adder{Base: 10}
	adder2 := Adder{Base: 20}

	fmt.Println("Expression with adder1:", addExpr(adder1, 5)) // 15
	fmt.Println("Expression with adder2:", addExpr(adder2, 5)) // 25
	// Output:
	// Expression with adder1: 15
	// Expression with adder2: 25

	// Methods as callbacks
	fmt.Println("\n--- Methods as Callbacks ---")
	type Handler struct {
		message string
	}

	handleClick := func(h *Handler) func() {
		return func() {
			fmt.Println("Clicked:", h.message)
		}
	}

	handler := &Handler{message: "Button 1"}
	callback := handleClick(handler)
	callback()
	// Output: Clicked: Button 1

	// When to use pointer vs value receivers
	fmt.Println("\n--- Receiver Type Guidelines ---")
	fmt.Println(`
Use POINTER receiver when:
  - Method needs to modify the receiver
  - Struct is large (avoids copying)
  - Consistency with other methods that use pointers
  - Struct contains sync.Mutex or similar

Use VALUE receiver when:
  - Method doesn't modify receiver
  - Struct is small (like Point{X, Y})
  - Value semantics are desired (immutability)
  - Basic type wrappers
`)
}
