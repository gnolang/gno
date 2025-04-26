package main

func main() {
	_ = &func() { println("Hello, World!") }
}

// Error:
// illegal assignment X expression type *gnolang.FuncLitExpr

// TypeCheckError:
// main/files/addressable_8a_err.gno:4:7: invalid operation: cannot take address of (func() literal) (value of type func())
