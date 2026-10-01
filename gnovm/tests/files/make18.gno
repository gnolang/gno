package main

func main() {
	l := -1
	_ = make([]int, l, 10)
}

// Error:
// runtime error: makeslice: len out of range
