package main

// Falling through from a clause that declares names into one that declares
// fewer (or none): all clauses share the switch's block, so the falling-through
// clause's names must be dropped first.
func main() {
	switch 1 {
	case 1:
		x := 1
		_ = x
		fallthrough
	case 2:
		println("two")
	}
	println("after")
}

// Output:
// two
// after
