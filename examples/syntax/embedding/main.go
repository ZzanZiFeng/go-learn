// Package main demonstrates struct embedding and composition in Go
package main

import (
	"fmt"
	"sync"
	"time"
)

// Base types for embedding examples

// Logger provides logging functionality
type Logger struct {
	prefix string
}

func (l Logger) Log(msg string) {
	fmt.Printf("[%s] LOG: %s\n", l.prefix, msg)
}

func (l Logger) Error(msg string) {
	fmt.Printf("[%s] ERROR: %s\n", l.prefix, msg)
}

// Metrics provides metrics tracking
type Metrics struct {
	requestCount int
}

func (m *Metrics) IncrementRequests() {
	m.requestCount++
}

func (m *Metrics) GetRequestCount() int {
	return m.requestCount
}

// Address represents a physical address
type Address struct {
	City    string
	Country string
}

func (a Address) FullAddress() string {
	return fmt.Sprintf("%s, %s", a.City, a.Country)
}

// Person is a base type
type Person struct {
	Name string
	Age  int
}

func (p Person) Introduce() string {
	return fmt.Sprintf("Hi, I'm %s, %d years old", p.Name, p.Age)
}

func (p Person) Greet() string {
	return "Hello!"
}

// Employee embeds Person and Address
type Employee struct {
	Person        // Anonymous embedding - fields promoted
	Address       // Anonymous embedding
	Title  string // Own field
	Salary float64
}

// Greet overrides Person.Greet
func (e Employee) Greet() string {
	return fmt.Sprintf("Hello, I'm %s, a %s", e.Name, e.Title)
}

// Service combines multiple mixins
type Service struct {
	Logger
	*Metrics // Embedded pointer
	name     string
}

func (s *Service) HandleRequest() {
	s.Log("Handling request")
	s.IncrementRequests()
	s.Log(fmt.Sprintf("Total requests: %d", s.GetRequestCount()))
}

// Animal demonstrates method overriding
type Animal struct {
	Name string
}

func (a Animal) Speak() string {
	return "..."
}

func (a Animal) Move() string {
	return a.Name + " moves"
}

// Dog embeds Animal and overrides Speak
type Dog struct {
	Animal
	Breed string
}

func (d Dog) Speak() string {
	return d.Name + " says: Woof!"
}

func (d Dog) Fetch() string {
	return d.Name + " fetches the ball"
}

// SafeCounter demonstrates embedding sync primitives
type SafeCounter struct {
	sync.Mutex
	value int
}

func (c *SafeCounter) Increment() {
	c.Lock()
	defer c.Unlock()
	c.value++
}

func (c *SafeCounter) Value() int {
	c.Lock()
	defer c.Unlock()
	return c.value
}

// Engine for composition example
type Engine struct {
	Power      int
	FuelType   string
	isRunning  bool
}

func (e *Engine) Start() {
	e.isRunning = true
	fmt.Printf("Engine started (Power: %d HP, Fuel: %s)\n", e.Power, e.FuelType)
}

func (e *Engine) Stop() {
	e.isRunning = false
	fmt.Println("Engine stopped")
}

func (e *Engine) IsRunning() bool {
	return e.isRunning
}

// Car composes Engine
type Car struct {
	*Engine // Embedded pointer for shared state
	Brand   string
	Model   string
}

func (c *Car) Drive() {
	if !c.IsRunning() {
		fmt.Println("Cannot drive - engine is not running!")
		return
	}
	fmt.Printf("%s %s is driving\n", c.Brand, c.Model)
}

// Interface embedding example
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

// CountingReader wraps a reader and counts bytes
type CountingReader struct {
	Reader    // Embedded interface
	BytesRead int
}

func (cr *CountingReader) Read(p []byte) (n int, err error) {
	n, err = cr.Reader.Read(p)
	cr.BytesRead += n
	return
}

// StringReader implements Reader
type StringReader struct {
	data string
	pos  int
}

func (sr *StringReader) Read(p []byte) (n int, err error) {
	if sr.pos >= len(sr.data) {
		return 0, fmt.Errorf("EOF")
	}
	n = copy(p, sr.data[sr.pos:])
	sr.pos += n
	return n, nil
}

// Timestamp mixin
type Timestamp struct {
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (t *Timestamp) Touch() {
	t.UpdatedAt = time.Now()
}

// Document uses Timestamp mixin
type Document struct {
	Timestamp
	Title   string
	Content string
}

func main() {
	fmt.Println("=== Embedding and Composition in Go ===")

	// Basic struct embedding
	fmt.Println("\n--- Basic Struct Embedding ---")
	emp := Employee{
		Person: Person{
			Name: "Alice",
			Age:  30,
		},
		Address: Address{
			City:    "Beijing",
			Country: "China",
		},
		Title:  "Engineer",
		Salary: 50000,
	}

	// Promoted fields - accessed directly
	fmt.Println("Name:", emp.Name)   // From Person
	fmt.Println("City:", emp.City)   // From Address
	fmt.Println("Title:", emp.Title) // Own field
	// Output:
	// Name: Alice
	// City: Beijing
	// Title: Engineer

	// Promoted methods
	fmt.Println("Introduction:", emp.Introduce()) // From Person
	fmt.Println("Full Address:", emp.FullAddress()) // From Address
	// Output:
	// Introduction: Hi, I'm Alice, 30 years old
	// Full Address: Beijing, China

	// Method overriding
	fmt.Println("\n--- Method Overriding ---")
	fmt.Println("Employee Greet:", emp.Greet())      // Overridden
	fmt.Println("Person Greet:", emp.Person.Greet()) // Original
	// Output:
	// Employee Greet: Hello, I'm Alice, a Engineer
	// Person Greet: Hello!

	// Mixin pattern
	fmt.Println("\n--- Mixin Pattern ---")
	svc := &Service{
		Logger:  Logger{prefix: "UserService"},
		Metrics: &Metrics{},
		name:    "user-service",
	}

	svc.HandleRequest()
	svc.HandleRequest()
	// Output:
	// [UserService] LOG: Handling request
	// [UserService] LOG: Total requests: 1
	// [UserService] LOG: Handling request
	// [UserService] LOG: Total requests: 2

	// Method override with Animal/Dog
	fmt.Println("\n--- Animal/Dog Example ---")
	dog := Dog{
		Animal: Animal{Name: "Buddy"},
		Breed:  "Golden Retriever",
	}

	fmt.Println(dog.Speak())        // Overridden
	fmt.Println(dog.Move())         // Inherited
	fmt.Println(dog.Fetch())        // Dog's own method
	fmt.Println(dog.Animal.Speak()) // Original Animal method
	// Output:
	// Buddy says: Woof!
	// Buddy moves
	// Buddy fetches the ball
	// ...

	// Embedding sync primitives
	fmt.Println("\n--- Embedding sync.Mutex ---")
	counter := &SafeCounter{}

	// Lock/Unlock are promoted
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			counter.Increment()
		}()
	}
	wg.Wait()
	fmt.Println("Counter value:", counter.Value())
	// Output: Counter value: 100

	// Composition with pointer embedding
	fmt.Println("\n--- Composition with Pointer ---")
	engine := &Engine{Power: 200, FuelType: "Electric"}
	car := &Car{
		Engine: engine,
		Brand:  "Tesla",
		Model:  "Model 3",
	}

	car.Drive() // Engine not running
	car.Start() // Promoted from Engine
	car.Drive()
	car.Stop()
	// Output:
	// Cannot drive - engine is not running!
	// Engine started (Power: 200 HP, Fuel: Electric)
	// Tesla Model 3 is driving
	// Engine stopped

	// Two cars sharing the same engine
	fmt.Println("\n--- Shared Embedded Pointer ---")
	car2 := &Car{
		Engine: engine, // Same engine as car
		Brand:  "Tesla",
		Model:  "Model Y",
	}

	car.Start()
	fmt.Println("Car engine running:", car.IsRunning())
	fmt.Println("Car2 engine running:", car2.IsRunning()) // Same engine!
	// Output:
	// Engine started (Power: 200 HP, Fuel: Electric)
	// Car engine running: true
	// Car2 engine running: true

	// Interface embedding - decorator pattern
	fmt.Println("\n--- Interface Embedding (Decorator) ---")
	sr := &StringReader{data: "Hello, World! This is a test."}
	cr := &CountingReader{Reader: sr}

	buf := make([]byte, 5)
	cr.Read(buf)
	fmt.Printf("Read: %q, Bytes so far: %d\n", string(buf), cr.BytesRead)
	cr.Read(buf)
	fmt.Printf("Read: %q, Bytes so far: %d\n", string(buf), cr.BytesRead)
	// Output:
	// Read: "Hello", Bytes so far: 5
	// Read: ", Wor", Bytes so far: 10

	// Timestamp mixin
	fmt.Println("\n--- Timestamp Mixin ---")
	doc := &Document{
		Timestamp: Timestamp{
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Title:   "My Document",
		Content: "Some content",
	}

	fmt.Println("Created:", doc.CreatedAt.Format("15:04:05"))
	time.Sleep(10 * time.Millisecond)
	doc.Touch()
	fmt.Println("Updated:", doc.UpdatedAt.Format("15:04:05.000"))
	// Output: timestamps differ slightly

	// Multiple embedding with name collision
	fmt.Println("\n--- Name Collision ---")
	type A struct{ Name string }
	type B struct{ Name string }
	type C struct {
		A
		B
		Name string // C's own Name shadows A and B
	}

	c := C{
		A:    A{Name: "from A"},
		B:    B{Name: "from B"},
		Name: "from C",
	}

	fmt.Println("c.Name:", c.Name)     // C's own
	fmt.Println("c.A.Name:", c.A.Name) // Explicitly A's
	fmt.Println("c.B.Name:", c.B.Name) // Explicitly B's
	// Output:
	// c.Name: from C
	// c.A.Name: from A
	// c.B.Name: from B

	// Summary
	fmt.Println("\n--- Composition vs Inheritance ---")
	fmt.Println(`
Go's approach: Composition over Inheritance

Benefits:
  ✓ Flexible - compose behaviors as needed
  ✓ Explicit - clear where methods come from
  ✓ No diamond problem
  ✓ Easy to change at runtime (with pointer embedding)

Patterns:
  - Mixins: Add functionality (Logger, Metrics)
  - Decorators: Wrap and extend (CountingReader)
  - Aggregation: Combine related data (Employee)
`)
}
