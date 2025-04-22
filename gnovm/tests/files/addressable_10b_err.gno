package main

type S struct {
	i int
}

func main() {
	println(&new(S))
}

// Error:
// illegal assignment X expression type *gnolang.CallExpr

// TypeCheckError:
// main/files/addressable_10b_err.gno:8:11: invalid operation: cannot take address of new(S) (value of type *S)
