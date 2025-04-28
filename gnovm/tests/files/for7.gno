package main

func main() {
	for i := 0; i; {
	}
}

// Error:
// main/files/for7.gno:4:2: expected typed bool kind, but got IntKind

// TypeCheckError:
// main/files/for7.gno:4:14: non-boolean condition in for statement
