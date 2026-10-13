package main

func main() {
	println(len(struct {
		A, B int
	}{}))
}

// Error:
// main/len8.gno:4:10-6:6: invalid argument: struct{A int; B int} for built-in len

// TypeCheckError:
// main/len8.gno:4:14: invalid argument: struct{A, B int}{} (value of type struct{A int; B int}) for built-in len
