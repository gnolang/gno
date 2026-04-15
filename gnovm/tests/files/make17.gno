package main

func main() {
	c := -1
	_ = make([]int, 0, c)
}

// Error:
// runtime error: makeslice: cap out of range
