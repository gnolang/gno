package main

// one untyped const, one typed const
func main() {
	println(1 + "a")
}

// Error:
// main/files/types/add_b2.gno:5:10: cannot use untyped Bigint as StringKind
