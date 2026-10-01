package main

func main() {
	defer func() {
		if r := recover(); r != nil {
			println("recovered")
		}

	}()
	l := 2
	_ = make([]int, l, 1)
}

// Output:
// recovered
