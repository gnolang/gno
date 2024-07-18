package main

func main() {
	var fns []func() int

	for i := 0; i < 3; i++ {
		z := i
		f := func() int {
			return z
		}
		fns = append(fns, f)
	}

	for _, fn := range fns {
		println(fn())
	}
}

// Output:
// 0
// 1
// 2
