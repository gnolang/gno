package main

const a = 1

func main() {
	_ = &a
}

// Error:
// main/dylan/addressable_5c_err.gno:6:6: cannot take address of (const (1 <untyped> bigint))
