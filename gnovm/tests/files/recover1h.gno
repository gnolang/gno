package main

import "fmt"

func main() {
	if r := recover(); r != nil { // Won't work
		fmt.Println("Recovered:", r)
	}
	panic("test panic")
}

// Error:
// panic: test panic
