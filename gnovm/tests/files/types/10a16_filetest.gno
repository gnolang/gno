package main

func main() {
	r := "hello" << 1
	println(r)
}

// Error:
// main/files/types/10a16_filetest.gno:3: operator << not defined on: <untyped> string
