package main

const a int = 1

func main() {
	_ = &a
}

// Error:
// main/addressable_5d_err.gno:6:6-8: invalid operation: cannot take address of (const (1 int)) (value of type int)

// TypeCheckError:
// main/addressable_5d_err.gno:6:7: invalid operation: cannot take address of a (constant 1 of type int)
