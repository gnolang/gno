package main

func main() {
	a := 1
	b := 1
	(a == b)++ // LHS is untyped bool, determined in preprocess
}

// Error:
// main/files/types2/shift_d44.gno:6: operator ++ not defined on: BoolKind
