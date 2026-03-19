package main

func main() {
	println(foo.Hello)
}

// Error:
// main/import15.gno:4:10-13: name foo not declared

// TypeCheckError:
// main/import15.gno:4:10: undefined: foo
