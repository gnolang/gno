package main

func main() {
	defer func() {
		r := recover()
		println("recover:", r)
	}()

	var p *int
	println(*p)
}

// Output:
// recover: nil pointer dereference
