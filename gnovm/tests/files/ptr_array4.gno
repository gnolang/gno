package main

func main() {
	defer func() {
		rec := recover()
		if rec != nil {
			println("recovered", rec)
		}
	}()
	var p *[3]int
	_ = p[0:1]
}

// Output:
// recovered nil pointer dereference
