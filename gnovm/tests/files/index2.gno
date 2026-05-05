package main

func main() {
	var a [1024]byte
	_ = a[-1]
}

// Error:
// main/index2.gno:5:6-11: invalid argument: index must not be negative: (-1 <untyped> bigint)

// TypeCheckError:
// main/index2.gno:5:8: invalid argument: index -1 (constant of type int) must not be negative
