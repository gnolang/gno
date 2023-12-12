package main

func main() {
	r := "hello" << 1
	println(r)
}

// Error:
// main/files/types/26a7_filetest.gno:4: operator << not defined on: <untyped> string
