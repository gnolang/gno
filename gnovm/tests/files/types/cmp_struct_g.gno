package main

import "fmt"

type Person struct {
	age int
}

type Dog struct {
	age int
}

func main() {
	a := Person{}
	b := Dog{}

	fmt.Println("a == b:", a == b)
}

// Error:
// main/cmp_struct_g.gno:17:25-31: invalid operation: (mismatched types main.Person and main.Dog)

// TypeCheckError:
// main/cmp_struct_g.gno:17:30: invalid operation: a == b (mismatched types Person and Dog)
