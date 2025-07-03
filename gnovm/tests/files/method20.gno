package main

import (
	"fmt"
)

type Hello struct{}

func (h *Hello) Hi() string {
	return "hi"
}

func main() {
	a := &Hello{}

	fmt.Println(a.Hi())
}

// Output:
// hi
