package main

func main() {
	println(len(struct {
		A, B int
	}{}))
}

// Error:
// unexpected type for len(): struct{A int; B int}

// TypeCheckError:
// main/files/len8.gno:4:14: invalid argument: struct{A, B int}{} (value of type struct{A int; B int}) for built-in len
