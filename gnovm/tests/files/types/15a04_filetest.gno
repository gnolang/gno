package main

func main() {
	a := 10.0
	a %= 3.0
	println(a)
}

// Error:
// main/files/types/15a04_filetest.gno:5: operator %= not defined on: Float64Kind
