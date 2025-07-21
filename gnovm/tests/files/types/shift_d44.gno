package main

func main() {
	a := 1
	b := 1
	(a == b)++ // LHS is untyped bool, determined in preprocess
}

// Error:
// main/shift_d44.gno:6:2-12: operator ++ not defined on: BoolKind

// TypeCheckError:
// main/shift_d44.gno:6:2: invalid operation: (a == b)++ (non-numeric type untyped bool)
