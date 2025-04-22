package main

const a = 1

func main() {
	_ = &a
}

// Error:
// illegal assignment X expression type *gnolang.ConstExpr

// TypeCheckError:
// main/files/addressable_5c_err.gno:6:7: invalid operation: cannot take address of a (untyped int constant 1)
