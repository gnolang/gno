package main

func main() {
	_, _ = &int(9)
}

// Error:
// main/files/addressable_7b_err.gno:4:9: cannot take address of (const (9 int))
