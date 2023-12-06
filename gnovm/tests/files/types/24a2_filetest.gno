package main

func main() {
	type Char byte
	var c Char = '世'
}

// Error:
// main/files/types/24a2_filetest.gno:3: rune overflows target kind
