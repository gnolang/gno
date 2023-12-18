package main

type integer int

func main() {
	println(integer(nil)) // TODO: this should be a conversion error: cannot convert nil to type integer
}

// Error:
// main/files/types/0f47a_filetest.gno:6: cannot convert (const (undefined) nil) to IntKind
