    package main

func main() {
	println(len(struct {
		A, B int
	}{}))
}

// Error:
// main/files/len8.gno:4:10: unexpected type for len(): struct{A int;B int}
