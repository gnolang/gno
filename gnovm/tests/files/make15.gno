package main

func main() {
	a := make([]int, 1, 1e20)
	println(len(a))
}

// Error:
// main/files/make15.gno:4:7: invalid capacity argument for built-in make: value out of range for int type
