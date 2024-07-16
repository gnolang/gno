package main

import "fmt"

type m map[string]int

func main() {
	// m(nil) is a conversion from nil to m, whose underlying type is map[string]int
	// lv: (nil main.m), rv: (undefined)
	if m(nil) == nil {
		fmt.Println("m is nil")
	} else {
		fmt.Println("m is not nil")
	}
}

// Output:
// m is nil
