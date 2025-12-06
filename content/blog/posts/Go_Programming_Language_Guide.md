<!-- 
{
    "tags": ["Go", "Instruction", "Example", "How-to"],
    "creation_date": "2025-12-03T14:00:00Z",
    "last_modification_date": "2025-12-04T16:00:00Z",
    "author": "Jason",
    "Description": "This is a description of the document and will override the first paragraph when looking at excerpts"
}
-->
# Go Programming Language Guide

## Introduction to Go

Go, also known as Golang, is a statically typed, compiled programming language designed at Google by Robert Griesemer, Rob Pike, and Ken Thompson. First released in 2009, Go has become one of the most popular languages for building modern, scalable applications.

### Why Choose Go?

Go offers several compelling features that make it an excellent choice for developers:

- **Fast compilation** - Go compiles to native machine code quickly
- **Built-in concurrency** - Goroutines and channels make concurrent programming simple
- **Strong standard library** - Comprehensive packages for common tasks
- **Simple syntax** - Easy to learn and read
- **Garbage collection** - Automatic memory management
- **Cross-platform** - Compile for multiple operating systems

## Basic Syntax

### Hello World

```go
package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}
```

### Variables and Types

Go supports various data types:

```go
// Variable declaration
var name string = "John"
var age int = 30
var isActive bool = true

// Short declaration
city := "New York"
temperature := 72.5

// Multiple declarations
var (
    username string = "admin"
    password string = "secret"
    attempts int = 3
)
```

## Control Structures

### If Statements

```go
if age >= 18 {
    fmt.Println("Adult")
} else if age >= 13 {
    fmt.Println("Teenager")
} else {
    fmt.Println("Child")
}
```

### For Loops

Go has only one looping construct - the `for` loop:

```go
// Traditional loop
for i := 0; i < 5; i++ {
    fmt.Println(i)
}

// While-style loop
count := 0
for count < 10 {
    count++
}

// Infinite loop
for {
    // Loop forever
    break // Exit with break
}

// Range loop
numbers := []int{1, 2, 3, 4, 5}
for index, value := range numbers {
    fmt.Printf("Index: %d, Value: %d\n", index, value)
}
```

### Switch Statements

```go
day := "Monday"

switch day {
case "Monday":
    fmt.Println("Start of the week")
case "Friday":
    fmt.Println("Almost weekend")
case "Saturday", "Sunday":
    fmt.Println("Weekend!")
default:
    fmt.Println("Midweek")
}
```

## Functions

### Basic Functions

```go
func add(a int, b int) int {
    return a + b
}

// Multiple return values
func divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, errors.New("division by zero")
    }
    return a / b, nil
}

// Named return values
func rectangle(width, height int) (area, perimeter int) {
    area = width * height
    perimeter = 2 * (width + height)
    return // Naked return
}
```

### Variadic Functions

```go
func sum(numbers ...int) int {
    total := 0
    for _, num := range numbers {
        total += num
    }
    return total
}

result := sum(1, 2, 3, 4, 5) // 15
```

## Data Structures

### Arrays and Slices

```go
// Array (fixed size)
var arr [5]int = [5]int{1, 2, 3, 4, 5}

// Slice (dynamic size)
slice := []string{"apple", "banana", "cherry"}
slice = append(slice, "date")

// Slice operations
subslice := slice[1:3] // ["banana", "cherry"]
length := len(slice)
capacity := cap(slice)
```

### Maps

```go
// Create a map
ages := make(map[string]int)
ages["Alice"] = 25
ages["Bob"] = 30

// Map literal
scores := map[string]int{
    "Alice": 95,
    "Bob":   87,
    "Carol": 92,
}

// Check if key exists
if age, exists := ages["Alice"]; exists {
    fmt.Println("Alice's age:", age)
}

// Delete key
delete(ages, "Bob")
```

### Structs

```go
type Person struct {
    FirstName string
    LastName  string
    Age       int
    Email     string
}

// Create struct instance
person := Person{
    FirstName: "John",
    LastName:  "Doe",
    Age:       30,
    Email:     "john@example.com",
}

// Access fields
fmt.Println(person.FirstName)

// Struct methods
func (p Person) FullName() string {
    return p.FirstName + " " + p.LastName
}
```

## Concurrency

### Goroutines

Goroutines are lightweight threads managed by the Go runtime:

```go
func sayHello(name string) {
    fmt.Printf("Hello, %s!\n", name)
}

// Launch goroutine
go sayHello("Alice")
go sayHello("Bob")

// Wait for goroutines
time.Sleep(time.Second)
```

### Channels

Channels are used for communication between goroutines:

```go
// Create channel
ch := make(chan int)

// Send to channel (in goroutine)
go func() {
    ch <- 42
}()

// Receive from channel
value := <-ch
fmt.Println(value)

// Buffered channel
buffered := make(chan string, 3)
buffered <- "first"
buffered <- "second"
buffered <- "third"
```

### Select Statement

```go
ch1 := make(chan string)
ch2 := make(chan string)

select {
case msg1 := <-ch1:
    fmt.Println("Received from ch1:", msg1)
case msg2 := <-ch2:
    fmt.Println("Received from ch2:", msg2)
case <-time.After(time.Second):
    fmt.Println("Timeout")
}
```

## Interfaces

Interfaces define behavior without implementation:

```go
type Shape interface {
    Area() float64
    Perimeter() float64
}

type Rectangle struct {
    Width  float64
    Height float64
}

func (r Rectangle) Area() float64 {
    return r.Width * r.Height
}

func (r Rectangle) Perimeter() float64 {
    return 2 * (r.Width + r.Height)
}

// Use interface
func printShapeInfo(s Shape) {
    fmt.Printf("Area: %.2f\n", s.Area())
    fmt.Printf("Perimeter: %.2f\n", s.Perimeter())
}
```

## Error Handling

Go uses explicit error returns rather than exceptions:

```go
func readFile(filename string) ([]byte, error) {
    data, err := os.ReadFile(filename)
    if err != nil {
        return nil, fmt.Errorf("failed to read file: %w", err)
    }
    return data, nil
}

// Custom error types
type ValidationError struct {
    Field string
    Issue string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Issue)
}
```

## Packages and Imports

### Creating a Package

```go
// File: math/calculator.go
package math

func Add(a, b int) int {
    return a + b
}

func Multiply(a, b int) int {
    return a * b
}
```

### Importing Packages

```go
import (
    "fmt"
    "net/http"
    "math/rand"
    
    "github.com/gorilla/mux"
    "myproject/math"
)
```

## Testing

Go has built-in testing support:

```go
// File: calculator_test.go
package math

import "testing"

func TestAdd(t *testing.T) {
    result := Add(2, 3)
    expected := 5
    
    if result != expected {
        t.Errorf("Add(2, 3) = %d; want %d", result, expected)
    }
}

func BenchmarkAdd(b *testing.B) {
    for i := 0; i < b.N; i++ {
        Add(2, 3)
    }
}
```

Run tests with:
```bash
go test
go test -v
go test -bench=.
```

## Common Use Cases

### Web Server

```go
package main

import (
    "fmt"
    "net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello, %s!", r.URL.Path[1:])
}

func main() {
    http.HandleFunc("/", handler)
    http.ListenAndServe(":8080", nil)
}
```

### JSON Handling

```go
import (
    "encoding/json"
    "fmt"
)

type User struct {
    Name  string `json:"name"`
    Email string `json:"email"`
    Age   int    `json:"age"`
}

// Marshal (struct to JSON)
user := User{Name: "Alice", Email: "alice@example.com", Age: 25}
jsonData, _ := json.Marshal(user)
fmt.Println(string(jsonData))

// Unmarshal (JSON to struct)
jsonString := `{"name":"Bob","email":"bob@example.com","age":30}`
var newUser User
json.Unmarshal([]byte(jsonString), &newUser)
```

## Best Practices

1. **Error handling** - Always check and handle errors explicitly
2. **Keep functions small** - Each function should do one thing well
3. **Use meaningful names** - Variables and functions should be descriptive
4. **Format code** - Use `go fmt` to maintain consistent formatting
5. **Write tests** - Test your code with the built-in testing package
6. **Document exports** - Add comments to exported functions and types
7. **Avoid global variables** - Pass dependencies explicitly
8. **Use interfaces** - Design for behavior, not implementation

## Useful Commands

| Command | Description |
|---------|-------------|
| `go run main.go` | Compile and run a Go program |
| `go build` | Compile packages and dependencies |
| `go test` | Run tests |
| `go fmt` | Format Go source code |
| `go get` | Download and install packages |
| `go mod init` | Initialize a new module |
| `go mod tidy` | Add missing and remove unused modules |
| `go install` | Compile and install packages |

## Conclusion

Go is a powerful, efficient language that excels at building:

- **Web services and APIs**
- **Cloud-native applications**
- **Command-line tools**
- **DevOps and site reliability tools**
- **Distributed systems**
- **Microservices**

Its simplicity, performance, and excellent concurrency support make it an ideal choice for modern software development.


> **Note:** This guide was generated using AI assistance in order to demonstrate markdown to HTML parsing

---

*For more information, visit the official Go documentation at [golang.org](https://golang.org)*