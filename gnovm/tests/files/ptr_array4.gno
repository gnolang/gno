package main

func main() {
	var p *[3]int
	_ = p[1:2]
}

// Error:
// runtime error: nil pointer dereference
