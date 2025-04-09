package main

func main() {
	_ = &func() { println("Hello, World!") }
}

// Error:
// main/files/addressable_8a_err.gno:4:6: illegal assignment X expression type *gnolang.FuncLitExpr
