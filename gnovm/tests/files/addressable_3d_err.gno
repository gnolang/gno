package main

func main() {
	_ = &(&struct{}{})
}

// Error:
// illegal assignment X expression type *gnolang.RefExpr

// TypeCheckError:
// main/files/addressable_3d_err.gno:4:7: invalid operation: cannot take address of (&struct{}{}) (value of type *struct{})
