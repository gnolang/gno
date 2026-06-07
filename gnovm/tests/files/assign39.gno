package main

// Verifies that in an AssignStmt, all RHS values are evaluated before LHS
// is written (Go spec §Assignments). This exercises the new interleaved
// Rhs[i]→Lhs[i] transcription order in transcribe.go.
func main() {
	x, y := 1, 2

	// Swap: both RHS (y and x) are evaluated before either LHS is written.
	x, y = y, x
	println(x, y) // 2 1

	// a=b (old), b=a+b (old a + old b)
	a, b := 3, 5
	a, b = b, a+b
	println(a, b) // 5 8
}

// Output:
// 2 1
// 5 8
