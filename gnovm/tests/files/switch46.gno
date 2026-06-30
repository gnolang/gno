package main

// Case expressions are evaluated in source order and short-circuit at the
// first match; a default in the middle is skipped during matching and does
// not run when a later case matches.
func sideEffect(s string) bool {
	println(s)
	return false
}

func main() {
	switch {
	case sideEffect("a"):
	default:
		println("default-should-not-run")
	case sideEffect("b"):
	case true:
		println("matched")
	case sideEffect("d-should-not-eval"):
	}
}

// Output:
// a
// b
// matched
