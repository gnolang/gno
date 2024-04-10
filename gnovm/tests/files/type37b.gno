package main

import "fmt"

type Integer int

func (i **Integer) Add(x int) {
	**i += Integer(x) // Dereference twice to get to the actual Integer value and modify it
}

func main() {
	a := new(Integer) // a is a pointer to Integer
	b := &a           // b is a pointer to a pointer to Integer

	// Since Add is defined on **Integer, you need to pass b
	b.Add(4) // Adds 4 to the value **b points to

	fmt.Println(**b) // Should print 4, as **b is the same as *a
}

// Error:
// main/files/type37b.gno:7: invalid receiver type **main.Integer (base type is pointer type)
