package main

func main() {
	true = false
}

// Error:
// main/assign_literal3.gno:4:2-14: cannot assign to uverse true

// TypeCheckError:
// main/assign_literal3.gno:4:2: cannot assign to true (neither addressable nor a map index expression)
