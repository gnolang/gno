package main

func main() {
	r := uint64(1<<2 == 1>>2) //special case with == !=, untyped bool
	println(r)
}

// Error:
// main/files/types/shift_d4.gno:4:7: cannot convert BoolKind to Uint64Kind

// TypeCheckError:
// main/files/types/shift_d4.gno:4:14: cannot convert 1 << 2 == 1 >> 2 (untyped bool constant false) to type uint64
