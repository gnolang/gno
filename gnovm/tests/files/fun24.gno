package main

func f(x int) (int, int) { return x, "foo" }

func main() {
	print("hello")
}

// Error:
// main/files/fun24.gno:3:28: cannot use untyped string as IntKind

// TypeCheckError:
// main/files/fun24.gno:3:38: cannot use "foo" (untyped string constant) as int value in return statement
