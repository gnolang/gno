package main

func main() {
	r := int(1)
	r &= int8(1)
	println(r)
}

// Error:
// main/files/types/18a2_filetest.gno:5: invalid operation: mismatched types int and int8
