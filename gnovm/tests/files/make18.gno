package main

func main() {
	l := -1
	_ = make([]int, l, 10)
}

// Error:
// makeslice: len out of range
