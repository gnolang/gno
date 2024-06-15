package main

func main() {
	println(X())
}

func X() string {
	return "hello" || "world"
}

// Error:
// main/files/bool6.gno:8:9: operands of boolean operators must evaluate to boolean typed values
