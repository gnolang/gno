package main

func main() {

	i := -2
	b := make([]byte, 1, i)

	println(b)
}

// Error:
// makeslice: cap out of range
