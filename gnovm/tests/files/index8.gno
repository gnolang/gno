package main

const c = -1

func main() {
	var a [5]int
	_ = a[c]
}

// Error:
// main/index8.gno:7:6-10: invalid argument: index must not be negative: (-1 <untyped> bigint)

// TypeCheckError:
// main/index8.gno:7:8: invalid argument: index c (constant -1 of type int) must not be negative
