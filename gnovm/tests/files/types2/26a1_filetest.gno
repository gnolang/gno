package main

func main() {
	a := 1
	b := 1
	(a == b)++ // LHS is untyped bool, determined in preprocess
}

// Error:
// main/files/types2/26a1_filetest.gno:6: operator ++ not defined on: BoolKind
