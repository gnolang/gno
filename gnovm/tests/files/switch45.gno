package main

// Chained fallthrough across three clauses, starting from a non-last default.
func main() {
	x := 9
	switch x {
	default:
		println("default")
		fallthrough
	case 1:
		println("one")
		fallthrough
	case 2:
		println("two")
	}
}

// Output:
// default
// one
// two
