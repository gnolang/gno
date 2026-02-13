package main

func main() {
	var a [1024]byte
	xx := a[-1]

	println(xx)
}

// Error:
// main/index_0.gno:5:8-13: invalid argument, index must not be negative: (-1 int)

// TypeCheckError:
// main/index_0.gno:5:10: invalid argument: index -1 (constant of type int) must not be negative
