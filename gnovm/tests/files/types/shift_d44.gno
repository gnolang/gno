package main

func main() {
	a := 1
	b := 1
	(a == b)++ // LHS is untyped bool, determined in preprocess
}

// Error:
// main/files/types/shift_d44.gno:6:2: operator ++ not defined on: BoolKind

// TypeCheckError:
// main/files/types/shift_d44.gno:6:2: invalid operation: (a == b)++ (non-numeric type untyped bool)
