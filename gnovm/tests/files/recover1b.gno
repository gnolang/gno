package main

import "fmt"

func main() {
	defer func() {
		recover()
		panic("other panic")
	}()
	panic("test panic")
}

// Stacktrace:
// panic other panic
// main()
//     main/files/recover1b.gno:8

// Error:
// other panic
