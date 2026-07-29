package main

func main() {
	l := -1
	_ = make([]int, l)
}

// Error:
// runtime error: makeslice: len out of range
