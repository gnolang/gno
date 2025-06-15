package main

func main() {
	var a int
	a, 2 = 6, 6
}

// Error:
// main/assign_literal2.gno:5:2-13: cannot assign to (const (2 <untyped> bigint))

// TypeCheckError:
// main/assign_literal2.gno:5:5: cannot assign to 2 (neither addressable nor a map index expression); main/assign_literal2.gno:4:6: declared and not used: a
