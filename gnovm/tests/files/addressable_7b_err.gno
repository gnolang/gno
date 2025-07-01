package main

func main() {
	_, _ = &int(9)
}

// Error:
// main/addressable_7b_err.gno:4:2-16: RHS should not be &((const (9 int))) when len(Lhs) > len(Rhs)

// TypeCheckError:
// main/addressable_7b_err.gno:4:10: invalid operation: cannot take address of int(9) (constant 9 of type int)
