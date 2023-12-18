package main

func main() {
	a := "hello"
	a++
	println(a)
}

// Error:
// main/files/types/25a1_filetest.gno:5: operator ++ not defined on: string
