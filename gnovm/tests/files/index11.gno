package main

func main() {
	_ = [5]int{-1: 5}
}

// Error:
// main/index11.gno:4:6-19: invalid argument: index must not be negative: (-1 <untyped> bigint)

// TypeCheckError:
// main/index11.gno:4:13: invalid argument: index -1 (constant of type int) must not be negative
