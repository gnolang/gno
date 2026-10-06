package main

func main() {
	var a [1024]byte
	i := -1
	_ = a[i]
}

// Error:
// runtime error: index out of range [-1]
