package main

import "fmt"

func shiftReturn() int64 {
	return 1<<4 + int(1) // The shift result is cast to int64
}

func main() {
	r := shiftReturn()

	fmt.Printf("%T, %d \n", r, r)
}

// Error:
// main/files/types/shift_d56.gno:6:2: cannot use int as int64
