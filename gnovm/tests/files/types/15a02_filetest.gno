package main

func main() {
	a := "hello"
	a %= 3
	println(a)
}

// Error:
// main/files/types/15a02_filetest.gno:5: operator %= not defined on: StringKind
