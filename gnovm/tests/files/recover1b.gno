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
// panic((const ("other panic" <untyped> string)))
//     main/files/recover1b.gno:8
// main()
//     main/files/recover1b.gno:5

// Error:
// other panic
