package main

func main() {
	println(&getPtr())
}

type S struct{}

func getPtr() *S {
	return &S{}
}

// Error:
// main/files/addressable_10a_err.gno:4:10: illegal assignment X expression type *gnolang.CallExpr
