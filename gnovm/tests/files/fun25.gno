package main

func f(x string) (a int, b int) { return x, 5 }

func main() {
	print("hello")
}

// Error:
// main/files/fun25.gno:3:35: cannot use string as int
