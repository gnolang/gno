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
// panic: other panic
// main<VPBlock(1,0)>()
//     main/files/recover1b.gno:8

// Error:
// other panic
