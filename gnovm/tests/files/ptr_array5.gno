package main

func main() {
	defer func() {
		rec := recover()
		if rec != nil {
			println("recovered", rec)
		}
	}()
	var p *[3]int
	for i, v := range p {
		println(i, v)
	}
}

// Output:
// recovered nil pointer dereference
