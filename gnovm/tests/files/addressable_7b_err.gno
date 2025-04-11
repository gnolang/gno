package main

func main() {
	_, _ = &int(9)
}

// Error:
// main/files/addressable_7b_err.gno:4:9: RHS should not be &((const (9 int))) when len(Lhs) > len(Rhs)
