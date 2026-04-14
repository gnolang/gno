package main

func main() {
	l := 2
	_ = make([]int, l, 1)
}

// Error:
// makeslice: cap out of range
