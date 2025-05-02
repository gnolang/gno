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
// defer func(){ ... }()
//     main/files/recover1b.gno:8
// main<VPBlock(1,0)>()
//     main/files/recover1b.gno:6

// Error:
// other panic

// TypeCheckError:
// main/files/recover1b.gno:3:8: "fmt" imported and not used
