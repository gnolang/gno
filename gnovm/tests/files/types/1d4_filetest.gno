package main

var a int
var b interface{}

func main() {
	println(b + a)
}

// Error:
// main/files/types/1d4_filetest.gno:7: invalid operation: mismatched types int and interface{}
