package main

func main() {
	x := -1
	println(1 >> x)
}

// Error:
// runtime error: negative shift amount: (-1 int)
