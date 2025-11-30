// Package main demonstrates struct basics in Go
package main

import (
	"encoding/json"
	"fmt"
)

// Person demonstrates basic struct definition
type Person struct {
	Name string // Exported field (public)
	Age  int    // Exported field (public)
	city string // Unexported field (private)
}

// Point demonstrates a simple struct with multiple fields
type Point struct {
	X, Y int
}

// Rectangle demonstrates struct with methods
type Rectangle struct {
	Width  float64
	Height float64
}

// Area calculates the area of the rectangle
func (r Rectangle) Area() float64 {
	return r.Width * r.Height
}

// User demonstrates struct with JSON tags
type User struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email,omitempty"`
	Password  string `json:"-"` // Ignored in JSON
}

// Address demonstrates embedded structs
type Address struct {
	City    string `json:"city"`
	Country string `json:"country"`
}

// Employee demonstrates struct embedding
type Employee struct {
	Name    string  `json:"name"`
	Title   string  `json:"title"`
	Address Address `json:"address"` // Named embedding
}

// Manager demonstrates anonymous embedding
type Manager struct {
	Employee        // Anonymous embedding - fields are promoted
	TeamSize int    `json:"team_size"`
}

func main() {
	fmt.Println("=== Struct Basics ===")

	// Different ways to initialize structs
	fmt.Println("\n--- Initialization Methods ---")

	// Method 1: Named fields (recommended)
	p1 := Person{
		Name: "Alice",
		Age:  30,
	}
	fmt.Printf("Named init: %+v\n", p1)
	// Output: Named init: {Name:Alice Age:30 city:}

	// Method 2: Positional initialization
	p2 := Person{"Bob", 25, "Beijing"}
	fmt.Printf("Positional init: %+v\n", p2)
	// Output: Positional init: {Name:Bob Age:25 city:Beijing}

	// Method 3: Zero value initialization
	var p3 Person
	fmt.Printf("Zero value init: %+v\n", p3)
	// Output: Zero value init: {Name: Age:0 city:}

	// Method 4: new() function - returns pointer
	p4 := new(Person)
	p4.Name = "Charlie"
	p4.Age = 35
	fmt.Printf("new() init: %+v\n", *p4)
	// Output: new() init: {Name:Charlie Age:35 city:}

	// Method 5: Address operator
	p5 := &Person{
		Name: "Diana",
		Age:  28,
	}
	fmt.Printf("Address operator: %+v\n", *p5)
	// Output: Address operator: {Name:Diana Age:28 city:}

	// Accessing fields
	fmt.Println("\n--- Field Access ---")
	fmt.Println("Name:", p1.Name)
	// Output: Name: Alice

	// Pointer auto-dereference
	fmt.Println("Pointer access:", p5.Name) // Same as (*p5).Name
	// Output: Pointer access: Diana

	// Modifying fields
	fmt.Println("\n--- Modifying Fields ---")
	p1.Age = 31
	fmt.Println("After modification:", p1.Age)
	// Output: After modification: 31

	// Struct methods
	fmt.Println("\n--- Struct Methods ---")
	rect := Rectangle{Width: 10, Height: 5}
	fmt.Println("Area:", rect.Area())
	// Output: Area: 50

	// JSON serialization
	fmt.Println("\n--- JSON Serialization ---")
	user := User{
		ID:        1,
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
		Password:  "secret123",
	}

	jsonData, _ := json.Marshal(user)
	fmt.Println("JSON:", string(jsonData))
	// Output: JSON: {"id":1,"first_name":"John","last_name":"Doe","email":"john@example.com"}
	// Note: Password is excluded due to json:"-"

	// JSON with omitempty
	userNoEmail := User{
		ID:        2,
		FirstName: "Jane",
		LastName:  "Smith",
	}
	jsonData2, _ := json.Marshal(userNoEmail)
	fmt.Println("JSON (omitempty):", string(jsonData2))
	// Output: JSON (omitempty): {"id":2,"first_name":"Jane","last_name":"Smith"}
	// Note: Email is omitted because it's empty and has omitempty tag

	// Struct embedding
	fmt.Println("\n--- Struct Embedding ---")
	emp := Employee{
		Name:  "Alice",
		Title: "Engineer",
		Address: Address{
			City:    "Beijing",
			Country: "China",
		},
	}
	empJSON, _ := json.MarshalIndent(emp, "", "  ")
	fmt.Println("Employee JSON:")
	fmt.Println(string(empJSON))
	// Output:
	// {
	//   "name": "Alice",
	//   "title": "Engineer",
	//   "address": {
	//     "city": "Beijing",
	//     "country": "China"
	//   }
	// }

	// Anonymous embedding with field promotion
	fmt.Println("\n--- Anonymous Embedding ---")
	mgr := Manager{
		Employee: Employee{
			Name:  "Bob",
			Title: "Tech Lead",
			Address: Address{
				City:    "Shanghai",
				Country: "China",
			},
		},
		TeamSize: 5,
	}

	// Promoted fields can be accessed directly
	fmt.Println("Manager Name:", mgr.Name) // Promoted from Employee
	fmt.Println("Manager City:", mgr.Address.City)
	// Output:
	// Manager Name: Bob
	// Manager City: Shanghai

	// Struct comparison
	fmt.Println("\n--- Struct Comparison ---")
	pt1 := Point{X: 1, Y: 2}
	pt2 := Point{X: 1, Y: 2}
	pt3 := Point{X: 2, Y: 3}

	fmt.Println("pt1 == pt2:", pt1 == pt2) // true
	fmt.Println("pt1 == pt3:", pt1 == pt3) // false

	// Anonymous structs
	fmt.Println("\n--- Anonymous Structs ---")
	config := struct {
		Host string
		Port int
	}{
		Host: "localhost",
		Port: 8080,
	}
	fmt.Printf("Config: %+v\n", config)
	// Output: Config: {Host:localhost Port:8080}

	// Table-driven test pattern with anonymous structs
	fmt.Println("\n--- Table-Driven Pattern ---")
	tests := []struct {
		input    int
		expected int
	}{
		{1, 2},
		{2, 4},
		{3, 6},
	}

	for _, tt := range tests {
		result := tt.input * 2
		if result == tt.expected {
			fmt.Printf("PASS: %d * 2 = %d\n", tt.input, result)
		}
	}
	// Output:
	// PASS: 1 * 2 = 2
	// PASS: 2 * 2 = 4
	// PASS: 3 * 2 = 6
}
