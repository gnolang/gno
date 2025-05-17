package main

const a int = 1

func main() {
	_ = &a
}

// Error:
// illegal assignment X expression type *gnolang.ConstExpr

// TypeCheckError:
// main/addressable_5d_err.gno:6:7: invalid operation: cannot take address of a (constant 1 of type int)
