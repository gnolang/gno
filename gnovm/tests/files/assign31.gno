package main

func main() {
	a := nil
	println(a)
}

// Error:
// main/files/assign31.gno:4:2: use of untyped nil in assignment

// TypeCheckError:
// main/files/assign31.gno:4:7: use of untyped nil in assignment
