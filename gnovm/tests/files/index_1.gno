package main

func main() {
	var a [1024]byte
	i := -1
	xx := a[i]

	println(xx)
}

// Error:
// runtime error: index out of range [-1]
