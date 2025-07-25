package main

func main() {
	a := [3]int{0, 1, 1: 1}
}

// Error:
// duplicate index 1 in array or slice literal

// TypeCheckError:
// main/array5.gno:4:20: duplicate index 1 in array or slice literal; main/array5.gno:4:2: declared and not used: a
