package main

// one untyped const, one typed const
func main() {
	println(1 & "a")
}

// Error:
// main/files/types/6b2_filetest.gno:5: operator & not defined on: <untyped> string
