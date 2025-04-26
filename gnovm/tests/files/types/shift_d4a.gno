package main

func main() {
	r := string(1<<2 == 1>>2) //special case with == !=, untyped bool
	println(r)
}

// Error:
// main/files/types/shift_d4a.gno:4:7: cannot convert BoolKind to StringKind

// TypeCheckError:
// main/files/types/shift_d4a.gno:4:14: cannot convert 1 << 2 == 1 >> 2 (untyped bool constant false) to type string
