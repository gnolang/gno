package main

func main() {
	var p *[3]int
	n := 0
	for range p {
		n++
	}
	println(n)
}

// Output:
// 3
