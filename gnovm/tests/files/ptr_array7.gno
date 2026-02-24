package main

func main() {
	defer func() {
		r := recover()
		println("recover:", r)
	}()

	var p *[3]int
	for range p {
	}
}

// Output:
// recover: nil pointer dereference
