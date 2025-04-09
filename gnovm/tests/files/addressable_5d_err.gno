package main

const a int = 1

func main() {
	_ = &a
}

// Error:
// main/files/addressable_5d_err.gno:6:6: illegal assignment X expression type *gnolang.ConstExpr
