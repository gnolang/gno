package main

func main() {
	t := make([]int, -1)
	println(t)
}

// Error:
// main/make3.gno:4:7-22: invalid argument: index (-1 <untyped> bigint) must not be negative

// TypeCheckError:
// main/make3.gno:4:19: invalid argument: index -1 (constant of type int) must not be negative
