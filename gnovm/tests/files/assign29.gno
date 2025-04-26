package main

func main() {
	t := struct{}
}

// Error:
// main/files/assign29.gno:4:2: struct{} (type) is not an expression

// TypeCheckError:
// main/files/assign29.gno:4:7: struct{} (type) is not an expression; main/files/assign29.gno:4:2: declared and not used: t
