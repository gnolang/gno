package main

func f(x int) (int, int) { return x, "foo" }

func main() {
	print("hello")
}

// Error:
// main/fun24.gno:3:28-43: cannot use untyped string as IntKind

// TypeCheckError:
// main/fun24.gno:3:38: cannot use "foo" (untyped string constant) as int value in return statement
