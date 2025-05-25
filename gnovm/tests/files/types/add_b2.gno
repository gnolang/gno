package main

// one untyped const, one typed const
func main() {
	println(1 + "a")
}

// Error:
// main/add_b2.gno:5:10-17: cannot use untyped Bigint as StringKind

// TypeCheckError:
// main/add_b2.gno:5:10: invalid operation: 1 + "a" (mismatched types untyped int and untyped string)
