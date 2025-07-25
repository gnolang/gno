package main

func main() {
	t := *struct{}
}

// Error:
// main/assign30.gno:4:2-16: *struct{} (type) is not an expression

// TypeCheckError:
// main/assign30.gno:4:7: *struct{} (type) is not an expression; main/assign30.gno:4:2: declared and not used: t
