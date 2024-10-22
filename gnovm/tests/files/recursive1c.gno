package main

import "fmt"

type S struct {
	A [2][2]S
}

func main() {
	var a, b S

	fmt.Println(a)
	fmt.Println(b)
}

// Error:
// main/files/recursive1c.gno:1:1: invalid recursive type: S -> S
