package main

func main() {
	1 + 2 = 6
}

// Error:
// main/files/types/assign_literal6.gno:4:2: cannot assign to (const (3 <untyped> bigint))

// TypeCheckError:
// main/files/types/assign_literal6.gno:4:2: cannot assign to 1 + 2 (neither addressable nor a map index expression)
