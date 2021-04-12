package main

func f(x string) (a int, b int) { return x, 5 }

func main() {
	print("hello")
}

// Error:
// cannot use string as int
