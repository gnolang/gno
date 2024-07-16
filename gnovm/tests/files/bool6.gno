package main

func main() {
	println(X())
}

func X() string {
	return "hello" || "world"
}

// Error:
// main/files/bool6.gno:8:9: operator || not defined on: StringKind
