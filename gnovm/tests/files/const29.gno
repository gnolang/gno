package main

import "fmt"

func main() {
	s := "1"
	const t = s
	fmt.Println(t)
}

// Error:
// main/const29.gno:7:8-13: s<VPBlock(1,0)> (variable of type string) is not constant

// TypeCheckError:
// main/const29.gno:7:12: s (variable of type string) is not constant
