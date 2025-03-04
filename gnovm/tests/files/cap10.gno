package main

func main() {
	println("cap", cap(struct{ A int }{}))
}

// Error:
// unexpected type for cap(): struct{A int}
