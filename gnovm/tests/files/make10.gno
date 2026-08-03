package main

func main() {
	t := make([]int, -1, 5)
	println(t)
}

// Error:
// main/make10.gno:4:7-25: invalid argument: index (-1 <untyped> bigint) must not be negative

// TypeCheckError:
// main/make10.gno:4:19: invalid argument: index -1 (constant of type int) must not be negative
