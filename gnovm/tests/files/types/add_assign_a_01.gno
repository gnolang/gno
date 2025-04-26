package main

func main() {
	a := 1
	b := 1
	a, b += 1, 1
	println(a)
	println(b)
}

// Error:
// main/files/types/add_assign_a_01.gno:6:2: assignment operator += requires only one expression on lhs and rhs

// TypeCheckError:
// main/files/types/add_assign_a_01.gno:6:7: assignment operation += requires single-valued expressions
