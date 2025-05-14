package main

const Pi = 3.14

func main() {
	Pi = 3.14159
	println(Pi)
}

// Error:
// main/files/types/assign_literal11.gno:6:2: cannot assign to const Pi

// TypeCheckError:
// main/files/types/assign_literal11.gno:6:2: cannot assign to Pi (neither addressable nor a map index expression)
