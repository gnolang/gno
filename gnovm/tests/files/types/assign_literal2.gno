package main

func main() {
	var a int
	a, 2 = 6, 6
}

// Error:
// main/files/types/assign_literal2.gno:5: cannot assign to (const (2 <untyped> bigint))
