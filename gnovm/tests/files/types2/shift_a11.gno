package main

func main() {
	a := 1
	r := a << "hello"
	println(r)
}

// Error:
// main/files/types2/shift_a11.gno:3: cannot convert StringKind to UintKind
