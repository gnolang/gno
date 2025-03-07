package main

import "fmt"

func main() {
	var err error

	switch v := err.(type) {
	case interface{ Format() string }:
		println("formatter")
	default:
		fmt.Println(v)
	}
}

// Output:
// <nil>
