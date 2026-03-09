package main

func main() {
	var p *[3]int
	_ = p[0:1]
}

// Error:
// nil pointer dereference
