package main

func main() {
	t := make(-1)
	println(t)
}

// Error:
// main/make12.gno:4:7-15: (const (-1 <untyped> bigint)) is not a type

// TypeCheckError:
// main/make12.gno:4:12: -1 is not a type
