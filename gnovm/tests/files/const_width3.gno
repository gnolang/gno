package main

// A literal wider than 512 bits is rejected too.

const c = 10000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000

func main() {
	println(c > 0)
}

// Error:
// main/const_width3.gno:5:11-169: constant overflow

// TypeCheckError:
// main/const_width3.gno:5:11: constant overflow
