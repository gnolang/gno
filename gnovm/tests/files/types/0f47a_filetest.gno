package main

type integer int

func main() {
	println(integer(nil)) // preprocess, 1042
}

// Error:
// main/files/types/0f47a_filetest.gno:6: cannot convert (const (undefined) nil) to IntKind
