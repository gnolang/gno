package main

func main() {
	i := -1
	t := make([]int, 5, i)
	println(t)
}

// Error:
// makeslice: cap out of range
