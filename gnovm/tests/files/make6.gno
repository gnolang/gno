package main

func main() {
	t := make([]int, 1, 1.2)
	println(t)
}

// TypeCheckError:
// main/make6.gno:4:22: 1.2 (untyped float constant) truncated to int

// Error:
// main/make6.gno:4:7-26: cannot convert untyped bigdec to integer -- 1.2 not an exact integer
