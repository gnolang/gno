package main

func main() {
	l := 2
	_ = make([]int, l, 1)
}

// Error:
// runtime error: makeslice: cap out of range
