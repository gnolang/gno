package main

import "fmt"

func shiftReturn() int64 {
	return 1 << 4 // The shift result is cast to int64
}

func main() {
	r := shiftReturn()

	fmt.Printf("%T, %d\n", r, r)
}

// Output:
// int64, 16
