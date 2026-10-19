package main

func main() {
	t := make([]int, 5, -1)
	println(t)
}

// Error:
// main/make11.gno:4:22-24: invalid argument: index -1 (constant of type int) must not be negative

// TypeCheckError:
// main/make11.gno:4:22: invalid argument: index -1 (constant of type int) must not be negative
