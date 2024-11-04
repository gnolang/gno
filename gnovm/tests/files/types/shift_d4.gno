package main

func main() {
	r := uint64(1<<2 == 1>>2) //special case with == !=, untyped bool
	println(r)
}

// Error:
// main/files/types/shift_d4.gno:4:7: cannot convert BoolKind to Uint64Kind
