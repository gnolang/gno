package main

func main() {
	var p *[3]int
	for i := range p {
		println(i)
	}
	println("ok")
}

// Output:
// 0
// 1
// 2
// ok
