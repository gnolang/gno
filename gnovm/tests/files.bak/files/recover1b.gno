package main

import "fmt"

func main() {
	defer func() {
		recover()
		panic("other panic")
	}()
	panic("test panic")
}

// Error:
// other panic
