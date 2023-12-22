package main

func main() {
	a := "hello"
	r := a << 1
	println(r)
}

// Error:
// main/files/types/10a14_filetest.gno:3: operator << not defined on: string
