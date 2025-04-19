package main

const a = 1

func main() {
	_ = &a
}

// Error:
// main/files/addressable_5c_err.gno:6:6: illegal assignment X expression type *gnolang.ConstExpr
