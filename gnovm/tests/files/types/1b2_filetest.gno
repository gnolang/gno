package main

// one untyped const, one typed const
func main() {
	println(1 + "a")
}

// Error:
// main/files/types/1b2_filetest.gno:5: cannot use untyped Bigint as StringKind
