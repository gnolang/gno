package main

func main() {
	a := nil
	println(a)
}

// Error:
// main/assign31.gno:4:2-10: use of untyped nil in assignment

// TypeCheckError:
// main/assign31.gno:4:7: use of untyped nil in assignment
