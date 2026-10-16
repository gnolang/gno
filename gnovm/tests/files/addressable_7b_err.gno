package main

func main() {
	_, _ = &int(9)
}

// Error:
// main/addressable_7b_err.gno:4:9-16: invalid operation: cannot take address of (const (9 int)) (value of type int)

// TypeCheckError:
// main/addressable_7b_err.gno:4:10: invalid operation: cannot take address of int(9) (constant 9 of type int)
