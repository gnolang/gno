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
// main/files/recursive1d.gno:1: type definition loop with S
