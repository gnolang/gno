package main

func main() {
	s := "hello"
	_ = s[-1]
}

// Error:
// main/index5.gno:5:6-11: invalid argument: index must not be negative: (-1 <untyped> bigint)

// TypeCheckError:
// main/index5.gno:5:8: invalid argument: index -1 (constant of type int) must not be negative
