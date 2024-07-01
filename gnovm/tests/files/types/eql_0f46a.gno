package main

import "fmt"

type s []int

func main() {
	// s(nil) is a conversion from nil to s, whose underlying type is []int
	// lv: (nil main.s), rv: (undefined)
	if s(nil) == nil {
		fmt.Println("s is nil")
	} else {
		fmt.Println("s is not nil")
	}
}

// Output:
// s is nil
