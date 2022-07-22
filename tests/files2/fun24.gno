package main

func f(x int) (int, int) { return x, "foo" }

func main() {
	print("hello")
}

// Error:
// main/files2/fun24.gno:3: cannot convert StringKind to IntKind