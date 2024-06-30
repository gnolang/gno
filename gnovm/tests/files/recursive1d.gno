package main

import "fmt"

type S struct {
	A [2]S
}

func main() {
	var a, b S

	fmt.Println(a)
	fmt.Println(b)
}

// Error:
// Cyclic dependency detected: S(File: files/recursive1d.gno, Line: 5) -> S(File: files/recursive1d.gno, Line: 5)
