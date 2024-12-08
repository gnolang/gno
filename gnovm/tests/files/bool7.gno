package main

// Ensure, when comparing evaluated boolean operand types, that the kinds produced
// are the same when one operand is typed and the other is untyped.
func main() {
	println(boolAndTrue(true))
	println(boolAndTrue(false))
}

func boolAndTrue(b bool) bool {
	return b && true
}

// Output:
// true
// false
