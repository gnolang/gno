package main

import "fmt"

// Define a Person struct
type Person struct {
	Name string
	Age  int
}

// Define an Employee struct that embeds Person
type Employee struct {
	Person
	Department string
	ID         int
}

func main() {
	emp1 := Employee{
		Person:     Person{Name: "John Doe", Age: 30},
		Department: "Engineering",
		ID:         12345,
	}

	emp2 := Employee{
		Person:     Person{Name: "John Doe", Age: 30},
		Department: "Engineering",
		ID:         12345,
	}

	emp3 := Employee{
		Person:     Person{Name: "Jane Doe", Age: 29},
		Department: "Marketing",
		ID:         67890,
	}

	fmt.Println("emp1 == emp2:", emp1 == emp2) // True because all fields match
	fmt.Println("emp1 == emp3:", emp1 == emp3) // False because some fields differ
}

// Output:
// emp1 == emp2: true
// emp1 == emp3: false
