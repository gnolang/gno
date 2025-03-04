package main

func main() {
	println(len(struct {
		A, B int
	}{}))
}

// Error:
// unexpected type for len(): struct{int, int}
