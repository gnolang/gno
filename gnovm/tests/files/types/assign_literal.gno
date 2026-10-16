package main

func main() {
	1 = 6
}

// Error:
// main/assign_literal.gno:4:2-7: cannot assign to 1

// TypeCheckError:
// main/assign_literal.gno:4:2: cannot assign to 1 (neither addressable nor a map index expression)
