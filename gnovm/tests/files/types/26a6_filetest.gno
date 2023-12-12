package main

func main() {
	a := "hello"
	r := a << 1
	println(r)
}

// Error:
// main/files/types/26a6_filetest.gno:5: operator << not defined on: string
