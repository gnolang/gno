package main

import (
	"fmt"
)

func foo() ([]string, interface{}) {
	return nil, fmt.Errorf("bar")
}

func main() {
	a, b := foo()
	fmt.Println(a, b)
}

// Output:
// [] bar
