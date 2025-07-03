package main

func main() {
	println(X())
}

func X() string {
	return "hello" || "world"
}

// Error:
// main/bool6.gno:8:9-27: operator || not defined on: StringKind

// TypeCheckError:
// main/bool6.gno:8:9: invalid operation: operator || not defined on "hello" (untyped string constant)
