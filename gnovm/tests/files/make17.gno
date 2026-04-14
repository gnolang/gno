package main

func main() {
	c := -1
	_ = make([]int, 0, c)
}

// Error:
// makeslice: cap out of range
