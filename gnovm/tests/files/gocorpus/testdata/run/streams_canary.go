// Verifies that the Go side is captured in program order across both
// streams: the builtin println writes to stderr in Go and to stdout in
// Gno, so a capture that buffered the streams separately would show
// Go's "B" last and report a divergence that the program does not have.

package main

import "fmt"

func main() {
	fmt.Println("A")
	println("B")
	fmt.Println("C")
}

// GnoOutput:
// A
// B
// C

// GoOutput:
// A
// B
// C

// Verdict: Clean run
