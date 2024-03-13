package main

func main() {
	s := []int{1, 2}

	f := func() {
		if len(s) == 2 {
			println("ok")
		}
	}
	f()
}

// Output:
// ok
