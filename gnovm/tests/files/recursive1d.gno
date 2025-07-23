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
// main/recursive1d.gno:5:6-7:2: invalid recursive type: S -> S

// TypeCheckError:
// main/recursive1d.gno:5:6: invalid recursive type: S refers to itself
