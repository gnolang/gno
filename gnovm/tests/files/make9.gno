package main

func main() {
	const i = -1
	t := make(map[int]string, i)
	println(t)
}

// Error:
// main/make9.gno:5:7-30: invalid argument: index (-1 <untyped> bigint) must not be negative

// TypeCheckError:
// main/make9.gno:5:28: invalid argument: index i (constant -1 of type int) must not be negative
