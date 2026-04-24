package main

func main() {
	const c1 = 1 < 8
	main()
	1
}

// Error:
// main/expr_stmt0.gno:6:2-3: invalid operation: (const (1 <untyped> bigint)) is not used

// TypeCheckError:
// main/expr_stmt0.gno:6:2: 1 (untyped int constant) is not used
