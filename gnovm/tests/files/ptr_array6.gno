package main

func main() {
	var p *[3]int
	for i, v := range p {
		println(i, v)
	}
}

// Error:
// nil pointer dereference
