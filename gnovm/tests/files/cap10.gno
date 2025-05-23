package main

func main() {
	println("cap", cap(struct{ A int }{}))
}

// Error:
// unexpected type for cap(): struct{A int}

// TypeCheckError:
// main/files/cap10.gno:4:21: invalid argument: struct{A int}{} (value of type struct{A int}) for built-in cap
