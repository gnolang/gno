package main

func main() {
	i := -1
	t := make([]int, 5, i)
	println(t)
}

// Error:
// runtime error: makeslice: cap out of range
