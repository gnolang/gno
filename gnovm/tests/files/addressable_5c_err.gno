package main

const a = 1

func main() {
	_ = &a
}

// Error:
// main/addressable_5c_err.gno:6:6-8: invalid operation: cannot take address of (const (1 <untyped> bigint)) (value of type <untyped> bigint)

// TypeCheckError:
// main/addressable_5c_err.gno:6:7: invalid operation: cannot take address of a (untyped int constant 1)
