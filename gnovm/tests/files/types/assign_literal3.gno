package main

func main() {
	true = false
}

// Error:
// main/files/types/assign_literal3.gno:4:2: cannot assign to uverse true

// TypeCheckError:
// main/files/types/assign_literal3.gno:4:2: cannot assign to true (neither addressable nor a map index expression)
