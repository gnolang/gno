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
// main/files/recursive1c.gno:1: type definition loop with S
