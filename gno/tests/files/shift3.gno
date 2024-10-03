package main

import "fmt"

type asciiSet [8]uint32

func main() {
	var as asciiSet
	var s string = "cx"
	var c rune = rune(s[0])
	as[c>>5] |= (1 << uint(c&31))
	fmt.Println(as)
}

// Output:
// [0 0 0 8 0 0 0 0]
