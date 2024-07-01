package main

func main() {
	r := uint64(1<<2 == 1>>2)
	println(r)
}

// Error:
// main/files/types2/shift_d4.gno:4: cannot convert BoolKind to Uint64Kind
