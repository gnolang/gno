package main

func f(x int) (int, int) { return x, "foo" }

func main() {
	print("hello")
}

// Error:
// main/files/fun24.gno:3:28: cannot convert StringKind to IntKind
