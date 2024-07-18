package main

func main() {
	s := []int{1, 2}

	f := func() {
		for i, v := range s {
			println(i)
			println(v)
		}
	}

	f()
}

// Output:
// 0
// 1
// 1
// 2
