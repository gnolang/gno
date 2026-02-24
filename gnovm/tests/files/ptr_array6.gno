package main

func main() {
	defer func() {
		r := recover()
		println("recover:", r)
	}()

	var p *[3]int
	_ = p[1:2]
}

// Output:
// recover: nil pointer dereference
