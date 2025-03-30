package main

func main() {
	println("cap", cap(struct{ A int }{}))
}

// Error:
// main/files/cap10.gno:4:17: unexpected type for cap(): struct{A int}
