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
// main/files/types/cmp_struct_g.gno:17:25: cannot use main.Person as main.Dog without explicit conversion

// TypeCheckError:
// main/files/types/cmp_struct_g.gno:17:30: invalid operation: a == b (mismatched types Person and Dog)
