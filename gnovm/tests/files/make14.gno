package main

func main() {
	t := make([]int, 6, 3)
	println(t)
}

// Error:
// main/make14.gno:4:7-24: invalid argument: len larger than cap in make([]int)

// TypeCheckError:
// main/make14.gno:4:19: invalid argument: length and capacity swapped
