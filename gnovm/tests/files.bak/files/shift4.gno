package main

type asciiSet [8]uint32

func main() {
	var as asciiSet
	var s string = "cx"
	var c rune = rune(s[0])
	println((as[c>>5] & (1 << uint(c&31))) != 0)
}

// Output:
// false
