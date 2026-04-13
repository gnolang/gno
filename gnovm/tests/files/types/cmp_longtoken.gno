package main

func main() {
	var a, b int

	// A very long, left-associative binary expression.
	// We are comparing the result (which is a boolean) to a string ("fail").
	// This triggers a compile-time type mismatch.
	_ = (((((((((((((((((((((a == b) ==
		(a == b)) ==
		(a == b)) ==
		(a == b)) ==
		(a == b)) ==
		(a == b)) ==
		(a == b)) ==
		(a == b)) ==
		(a == b)) ==
		(a == b)) ==
		(a == b)) ==
		(a == b)) ==
		(a == b)) ==
		(a == b)) ==
		(a == b)) ==
		(a == b)) ==
		(a == b)) ==
		(a == b)) ==
		(a == b)) ==
		(a == b)) == "fail")
}

// Error:
// main/cmp_longtoken.gno:9:7-28:22: invalid operation: (mismatched types <untyped> bool and <untyped> string)

// TypeCheckError:
// main/cmp_longtoken.gno:28:16: invalid operation: ((((((((((((((((((((a == b) == (a == b)) == (a == b)) == (a == b)) == (a == b)) == (a == b)) == (a == b)) == (a == b)) == (a == b)) == (a == b)) == (a == b)) == (a == b)) == (a == b)) == (a == b)) == (a == b)) == (a == b)) == (a == b)) == (a == b)) == (a == b)) == (a == b)) == "fail" (mismatched types untyped bool and untyped string)
