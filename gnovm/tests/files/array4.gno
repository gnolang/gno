package main

func main() {
	a := [3]int{2: 2, 1: 1, 2: 2}
}

// Error:
// duplicate index 2 in array or slice literal

// TypeCheckError:
// main/files/array4.gno:4:26: duplicate index 2 in array or slice literal; main/files/array4.gno:4:2: declared and not used: a
