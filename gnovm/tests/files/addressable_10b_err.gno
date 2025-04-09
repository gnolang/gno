package main

type S struct {
	i int
}

func main() {
	println(&new(S))
}

// Error:
// main/files/addressable_10b_err.gno:8:10: illegal assignment X expression type *gnolang.CallExpr
