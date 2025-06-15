package main

func main() {
	a := [2]int{1, 2}

	const b = a
}

// Error:
// main/const43.gno:6:8-13: a<VPBlock(1,0)> (variable of type [2]int) is not constant

// TypeCheckError:
// main/const43.gno:6:12: a (variable of type [2]int) is not constant
