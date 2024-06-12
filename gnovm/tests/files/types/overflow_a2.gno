package main

func main() {
	type Char byte
	var c Char = '世'
}

// Error:
// main/files/types/overflow_a2.gno:3: rune overflows target kind
