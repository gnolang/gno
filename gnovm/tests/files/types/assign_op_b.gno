package main

func main() {
	s := "abc"
	s[0] += 1
	println("ok")
}

// Error:
// main/assign_op_b.gno:5:2-11: cannot assign to s[0] (neither addressable nor a map index expression)

// TypeCheckError:
// main/assign_op_b.gno:5:2: cannot assign to s[0] (neither addressable nor a map index expression)
