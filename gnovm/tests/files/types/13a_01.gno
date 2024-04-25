package main

func main() {
	a := 1
	b := 1
	a, b += 1, 1
	println(a)
	println(b)
}

// Error:
// main/files/types/13a_01.gno:6: assignment operator += requires only one expression on lhs and rhs
