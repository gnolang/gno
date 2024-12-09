package main

func main() {
	a := [2]int{1, 2}

	const b = a
}

// Error:
// main/files/const43.gno:6:8: a (variable of type [2]int) is not constant
