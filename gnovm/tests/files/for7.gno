package main

func main() {
	for i := 0; i; {
	}
}

// Error:
// main/for7.gno:4:2-5:3: expected typed bool kind, but got IntKind

// TypeCheckError:
// main/for7.gno:4:14: non-boolean condition in for statement
