package main

func main() {
	var p *[3]int
	for range p {
	}
}

// Error:
// nil pointer dereference
