package main

func main() {
	a := 1
	r := a << "hello"
	println(r)
}

// Error:
// main/files/types2/shift_a11.gno:3:1: cannot convert StringKind to UintKind
