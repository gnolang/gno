package main

func main() {
	_ = &nil
}

// TypeCheckError:
// main/addressable_12d_err.gno:4:7: invalid operation: cannot take address of nil

// Error:
// main/addressable_12d_err.gno:4:6-10: invalid operation: cannot take address of (const (undefined))
