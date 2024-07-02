package main

import "fmt"

type S struct {
	A [2][2]S
}

func main() {
	var a, b S

	//println(a == b)
	fmt.Println(a)
	fmt.Println(b)
}

// Error:
// Cyclic dependency detected: S(File: files/recursive1c.gno) -> S(File: files/recursive1c.gno)
