package main

func main() {
	r := uint64(1<<2 == 1>>2) //special case with == !=, untyped bool
	println(r)
}

// Error:
// main/files/types2/shift_d4.gno:4:7: cannot convert (const (false <untyped> bool)) to integer type
