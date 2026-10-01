package main

func main() {
	_ = []int{-1: 5}
}

// Error:
// main/index10.gno:4:6-18: invalid argument: index must not be negative: (-1 <untyped> bigint)

// TypeCheckError:
// main/index10.gno:4:12: invalid argument: index -1 (constant of type int) must not be negative
