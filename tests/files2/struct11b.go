package main

import (
	"fmt"
)

type ResponseWriter interface {
	FooBar()
}

type Fromage struct {
	ResponseWriter
}

func main() {
	a := Fromage{}
	if a.ResponseWriter == nil {
		fmt.Println("nil")
	} else {
		fmt.Println("not nil")
	}
}

// Output:
// nil
