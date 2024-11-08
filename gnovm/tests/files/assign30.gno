package main

func main() {
	t := *struct{}
}

// Error:
// main/files/assign30.gno:4:2: *struct{} (type) is not an expression
