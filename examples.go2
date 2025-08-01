// Every Go program starts with a 'package' declaration.
// 'main' is the starting point of execution.
package main

// 'fmt' is the standard library for formatted I/O (printing output)
import (
	"errors"
	"fmt"
	"time"
)

// Declare a constant
const Pi = 3.14

// Define a struct (like a class in OOP)
type Person struct {
	Name string
	Age  int
}

// Define an interface (a contract for behavior)
type Speaker interface {
	Speak()
}

// Implementing the 'Speaker' interface for 'Person'
func (p Person) Speak() {
	fmt.Printf("Hello, my name is %s and I am %d years old.\n", p.Name, p.Age)
}

// Function that returns sum of two integers
func add(a int, b int) int {
	return a + b
}

// Function that returns multiple values
func divide(a int, b int) (int, error) {
	if b == 0 {
		return 0, errors.New("cannot divide by zero")
	}
	return a / b, nil
}

// Function to demonstrate goroutine (lightweight thread)
func printCount(label string) {
	for i := 1; i <= 5; i++ {
		fmt.Printf("%s: %d\n", label, i)
		time.Sleep(500 * time.Millisecond) // Sleep for 0.5 sec
	}
}

// Entry point of the program
func main() {
	// Variables and basic data types
	var name string = "GoLang"
	age := 15 // short-hand declaration with type inference
	fmt.Println("Welcome to", name)
	fmt.Println("Age is", age)

	// Constants
	fmt.Println("Value of Pi:", Pi)

	// If-else condition
	if age > 10 {
		fmt.Println("Age is greater than 10")
	} else {
		fmt.Println("Age is 10 or less")
	}

	// Switch-case
	grade := "B"
	switch grade {
	case "A":
		fmt.Println("Excellent!")
	case "B":
		fmt.Println("Good Job!")
	default:
		fmt.Println("Keep Trying!")
	}

	// Loops - 'for' loop (no 'while' in Go)
	for i := 1; i <= 5; i++ {
		fmt.Println("Count:", i)
	}

	// Arrays - fixed size collection
	var arr = [3]int{10, 20, 30}
	fmt.Println("Array elements:", arr)

	// Slices - dynamically-sized array
	slice := []int{1, 2, 3}
	slice = append(slice, 4)
	fmt.Println("Slice elements:", slice)

	// Maps - key-value pairs (like dictionaries)
	personAge := map[string]int{
		"Alice": 25,
		"Bob":   30,
	}
	personAge["Charlie"] = 35
	fmt.Println("Person Age Map:", personAge)
	fmt.Println("Age of Bob:", personAge["Bob"])

	// Struct usage
	p := Person{Name: "Naresh", Age: 28}
	p.Speak()

	// Pointers - variables that hold memory addresses
	x := 10
	var ptr *int = &x
	fmt.Println("Value of x:", x)
	fmt.Println("Address of x:", ptr)
	fmt.Println("Value from pointer:", *ptr)

	// Function calls
	sum := add(5, 3)
	fmt.Println("Sum:", sum)

	// Handling multiple return values and errors
	result, err := divide(10, 0)
	if err != nil {
		fmt.Println("Error occurred:", err)
	} else {
		fmt.Println("Division result:", result)
	}

	// Goroutines - light-weight concurrency
	fmt.Println("Starting Goroutines...")
	go printCount("Goroutine 1")
	go printCount("Goroutine 2")

	// Wait for goroutines to finish (basic way)
	time.Sleep(3 * time.Second)

	fmt.Println("Program Ended.")
}
