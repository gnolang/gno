package main

func main() {
	var a = []int{1, 2, 3}
	_ = a[-1]
}

// Error:
// main/index4.gno:5:6-11: invalid argument: index must not be negative: (-1 <untyped> bigint)

// TypeCheckError:
// main/index4.gno:5:8: invalid argument: index -1 (constant of type int) must not be negative
