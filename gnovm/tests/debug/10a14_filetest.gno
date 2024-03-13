package main

func main() {
	a := "hello"
	r := a << 1
	println(r)
}

// Error:
// main/debug/10a14_filetest.gno:5: operator << not defined on: StringKind
