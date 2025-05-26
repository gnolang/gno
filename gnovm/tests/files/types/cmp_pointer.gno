package main

import "fmt"

// Define a Person struct
type Person struct {
	Name string
	Age  int
}

type Worker string

func main() {
	// Create two pointers to Person instances
	p1 := &Person{Name: "Alice", Age: 30}
	p2 := Worker("a")
	p2Ptr := &p2

	fmt.Println("p1 == p2:", p1 == p2Ptr)
}

// Error:
// main/cmp_pointer.gno:19:27-38: cannot use main.Person as main.Worker without explicit conversion

// TypeCheckError:
// main/cmp_pointer.gno:19:33: invalid operation: p1 == p2Ptr (mismatched types *Person and *Worker)
