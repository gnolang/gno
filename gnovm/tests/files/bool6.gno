package main

func main() {
	println(X())
}

func X() string {
	return "hello" || "world"
}

// Error:
// main/files/bool6.gno:8:9: operator || not defined on: StringKind

// TypeCheckError:
// main/files/bool6.gno:8:9: invalid operation: operator || not defined on "hello" (untyped string constant)
