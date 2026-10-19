package main

func main() {
	println("cap", cap(struct{ A int }{}))
}

// Error:
// main/cap10.gno:4:17-39: invalid argument: struct{A int} for built-in cap

// TypeCheckError:
// main/cap10.gno:4:21: invalid argument: struct{A int}{} (value of type struct{A int}) for built-in cap
