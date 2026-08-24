package main

func main() {
	t := make([]int, 1, -1.0)
	println(t)
}

// Error:
// main/make4.gno:4:7-27: invalid argument: index (-1.0 <untyped> bigdec) must not be negative

// TypeCheckError:
// main/make4.gno:4:22: invalid argument: index -1.0 (constant -1 of type int) must not be negative
