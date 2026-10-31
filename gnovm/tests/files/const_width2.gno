package main

// The width limit applies to every operation, named by the operator as
// go/types does.

const c = 1 << 511

func main() {
	println(c + c)
}

// Error:
// main/const_width2.gno:9:10-15: constant addition overflow

// TypeCheckError:
// main/const_width2.gno:9:12: constant addition overflow
