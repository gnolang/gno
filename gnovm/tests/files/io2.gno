package main

import (
	"fmt"
	"io"
	"strings"
)

func main() {
	r := strings.NewReader("Go is a general-purpose language designed with systems programming in mind.")

	b, err := io.ReadAll(r)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", b)
}

// Output:
// Go is a general-purpose language designed with systems programming in mind.
