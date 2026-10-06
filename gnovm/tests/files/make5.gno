package main

func main() {
	t := make([]int, 1, -1.2)
	println(t)
}

// Error:
// main/make5.gno:4:7-27: invalid argument: index (-1.2 <untyped> bigdec) must not be negative

// TypeCheckError:
// main/make5.gno:4:22: -1.2 (untyped float constant) truncated to int
