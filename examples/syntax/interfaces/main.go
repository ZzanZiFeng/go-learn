// Package main demonstrates interfaces in Go
package main

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

// Speaker interface demonstrates basic interface definition
type Speaker interface {
	Speak() string
}

// Dog implements Speaker implicitly
type Dog struct {
	Name string
}

func (d Dog) Speak() string {
	return d.Name + " says: Woof!"
}

// Cat implements Speaker implicitly
type Cat struct {
	Name string
}

func (c Cat) Speak() string {
	return c.Name + " says: Meow!"
}

// Robot implements Speaker implicitly
type Robot struct {
	Model string
}

func (r Robot) Speak() string {
	return "Robot " + r.Model + " says: Beep boop!"
}

// Shape interface for polymorphism
type Shape interface {
	Area() float64
	Perimeter() float64
}

// Rectangle implements Shape
type Rectangle struct {
	Width, Height float64
}

func (r Rectangle) Area() float64 {
	return r.Width * r.Height
}

func (r Rectangle) Perimeter() float64 {
	return 2 * (r.Width + r.Height)
}

// Circle implements Shape
type Circle struct {
	Radius float64
}

func (c Circle) Area() float64 {
	return math.Pi * c.Radius * c.Radius
}

func (c Circle) Perimeter() float64 {
	return 2 * math.Pi * c.Radius
}

// Triangle implements Shape
type Triangle struct {
	A, B, C float64 // sides
}

func (t Triangle) Area() float64 {
	// Heron's formula
	s := (t.A + t.B + t.C) / 2
	return math.Sqrt(s * (s - t.A) * (s - t.B) * (s - t.C))
}

func (t Triangle) Perimeter() float64 {
	return t.A + t.B + t.C
}

// Stringer interface (from fmt package)
type Person struct {
	Name string
	Age  int
}

func (p Person) String() string {
	return fmt.Sprintf("%s (%d years old)", p.Name, p.Age)
}

// Interface composition example
type Reader interface {
	Read(p []byte) (n int, err error)
}

type Writer interface {
	Write(p []byte) (n int, err error)
}

type ReadWriter interface {
	Reader
	Writer
}

// Buffer implements ReadWriter
type Buffer struct {
	data []byte
	pos  int
}

func (b *Buffer) Read(p []byte) (n int, err error) {
	if b.pos >= len(b.data) {
		return 0, fmt.Errorf("EOF")
	}
	n = copy(p, b.data[b.pos:])
	b.pos += n
	return n, nil
}

func (b *Buffer) Write(p []byte) (n int, err error) {
	b.data = append(b.data, p...)
	return len(p), nil
}

// Empty interface examples
func PrintAnything(v interface{}) {
	fmt.Printf("Type: %T, Value: %v\n", v, v)
}

// Type assertion examples
func DescribeValue(i interface{}) {
	switch v := i.(type) {
	case int:
		fmt.Printf("Integer: %d (doubled: %d)\n", v, v*2)
	case string:
		fmt.Printf("String: %q (length: %d)\n", v, len(v))
	case bool:
		fmt.Printf("Boolean: %t\n", v)
	case []int:
		fmt.Printf("Int slice with %d elements: %v\n", len(v), v)
	case Shape:
		fmt.Printf("Shape with area: %.2f\n", v.Area())
	default:
		fmt.Printf("Unknown type: %T\n", v)
	}
}

// ByAge implements sort.Interface for []Person
type ByAge []Person

func (a ByAge) Len() int           { return len(a) }
func (a ByAge) Less(i, j int) bool { return a[i].Age < a[j].Age }
func (a ByAge) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func main() {
	fmt.Println("=== Interfaces in Go ===")

	// Basic interface usage
	fmt.Println("\n--- Basic Interface Usage ---")
	var speaker Speaker

	speaker = Dog{Name: "Buddy"}
	fmt.Println(speaker.Speak())
	// Output: Buddy says: Woof!

	speaker = Cat{Name: "Whiskers"}
	fmt.Println(speaker.Speak())
	// Output: Whiskers says: Meow!

	speaker = Robot{Model: "R2D2"}
	fmt.Println(speaker.Speak())
	// Output: Robot R2D2 says: Beep boop!

	// Interface slice - polymorphism
	fmt.Println("\n--- Polymorphism with Interface Slice ---")
	speakers := []Speaker{
		Dog{Name: "Max"},
		Cat{Name: "Luna"},
		Robot{Model: "C3PO"},
	}

	for _, s := range speakers {
		fmt.Println(s.Speak())
	}
	// Output:
	// Max says: Woof!
	// Luna says: Meow!
	// Robot C3PO says: Beep boop!

	// Shape interface
	fmt.Println("\n--- Shape Interface ---")
	shapes := []Shape{
		Rectangle{Width: 10, Height: 5},
		Circle{Radius: 3},
		Triangle{A: 3, B: 4, C: 5},
	}

	for _, shape := range shapes {
		fmt.Printf("%T: Area=%.2f, Perimeter=%.2f\n",
			shape, shape.Area(), shape.Perimeter())
	}
	// Output:
	// main.Rectangle: Area=50.00, Perimeter=30.00
	// main.Circle: Area=28.27, Perimeter=18.85
	// main.Triangle: Area=6.00, Perimeter=12.00

	// Calculate total area
	totalArea := 0.0
	for _, shape := range shapes {
		totalArea += shape.Area()
	}
	fmt.Printf("Total area: %.2f\n", totalArea)
	// Output: Total area: 84.27

	// fmt.Stringer interface
	fmt.Println("\n--- fmt.Stringer Interface ---")
	person := Person{Name: "Alice", Age: 30}
	fmt.Println(person) // String() is called automatically
	// Output: Alice (30 years old)

	// Empty interface
	fmt.Println("\n--- Empty Interface (interface{}) ---")
	PrintAnything(42)
	PrintAnything("hello")
	PrintAnything(true)
	PrintAnything([]int{1, 2, 3})
	PrintAnything(Rectangle{Width: 5, Height: 3})
	// Output:
	// Type: int, Value: 42
	// Type: string, Value: hello
	// Type: bool, Value: true
	// Type: []int, Value: [1 2 3]
	// Type: main.Rectangle, Value: {5 3}

	// Type assertion with comma-ok
	fmt.Println("\n--- Type Assertion (comma-ok) ---")
	var i interface{} = "hello"

	// Safe type assertion
	s, ok := i.(string)
	if ok {
		fmt.Println("String value:", strings.ToUpper(s))
	}
	// Output: String value: HELLO

	// Failed assertion returns zero value
	n, ok := i.(int)
	fmt.Printf("Int assertion: value=%d, ok=%t\n", n, ok)
	// Output: Int assertion: value=0, ok=false

	// Type switch
	fmt.Println("\n--- Type Switch ---")
	values := []interface{}{
		42,
		"hello",
		true,
		[]int{1, 2, 3},
		Circle{Radius: 2},
		3.14,
	}

	for _, v := range values {
		DescribeValue(v)
	}
	// Output:
	// Integer: 42 (doubled: 84)
	// String: "hello" (length: 5)
	// Boolean: true
	// Int slice with 3 elements: [1 2 3]
	// Shape with area: 12.57
	// Unknown type: float64

	// Interface composition
	fmt.Println("\n--- Interface Composition ---")
	buf := &Buffer{}
	buf.Write([]byte("Hello, World!"))

	data := make([]byte, 5)
	n, _ := buf.Read(data)
	fmt.Printf("Read %d bytes: %s\n", n, string(data))
	// Output: Read 5 bytes: Hello

	// Verify Buffer implements ReadWriter
	var rw ReadWriter = buf
	rw.Write([]byte(" More data"))
	fmt.Printf("Buffer data: %s\n", string(buf.data))
	// Output: Buffer data: Hello, World! More data

	// sort.Interface example
	fmt.Println("\n--- sort.Interface Example ---")
	people := []Person{
		{Name: "Alice", Age: 30},
		{Name: "Bob", Age: 25},
		{Name: "Charlie", Age: 35},
	}

	fmt.Println("Before sorting:")
	for _, p := range people {
		fmt.Printf("  %s: %d\n", p.Name, p.Age)
	}

	sort.Sort(ByAge(people))

	fmt.Println("After sorting by age:")
	for _, p := range people {
		fmt.Printf("  %s: %d\n", p.Name, p.Age)
	}
	// Output:
	// Before sorting:
	//   Alice: 30
	//   Bob: 25
	//   Charlie: 35
	// After sorting by age:
	//   Bob: 25
	//   Alice: 30
	//   Charlie: 35

	// Compile-time interface check
	fmt.Println("\n--- Compile-time Interface Check ---")
	// This line will cause a compile error if Dog doesn't implement Speaker
	var _ Speaker = Dog{}
	var _ Speaker = (*Cat)(nil) // Check pointer implements interface
	fmt.Println("All types implement their interfaces correctly!")

	// Interface nil check
	fmt.Println("\n--- Interface Nil Check ---")
	var nilSpeaker Speaker
	fmt.Printf("Nil interface: %v, is nil: %t\n", nilSpeaker, nilSpeaker == nil)
	// Output: Nil interface: <nil>, is nil: true

	// Interface value structure
	fmt.Println("\n--- Interface Value Structure ---")
	var iface interface{}
	fmt.Printf("Empty: type=%T, value=%v, nil=%t\n", iface, iface, iface == nil)

	iface = 42
	fmt.Printf("Int: type=%T, value=%v, nil=%t\n", iface, iface, iface == nil)

	iface = "hello"
	fmt.Printf("String: type=%T, value=%v, nil=%t\n", iface, iface, iface == nil)
	// Output:
	// Empty: type=<nil>, value=<nil>, nil=true
	// Int: type=int, value=42, nil=false
	// String: type=string, value=hello, nil=false
}
