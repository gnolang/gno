package main

const c int = 5

func main() {
	c += 1
	println("ok")
}

// Error:
// main/assign_op_a.gno:6:2-8: cannot assign to const c

// TypeCheckError:
// main/assign_op_a.gno:6:2: cannot assign to c (neither addressable nor a map index expression)
