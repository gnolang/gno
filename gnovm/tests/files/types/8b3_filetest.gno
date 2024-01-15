package main

// one untyped const, one typed const
func main() {
	println("b" ^ "a")
}

// Error:
// main/files/types/8b3_filetest.gno:5: operator ^ not defined on: StringKind
