package main

func main() {
	var fns []func() int

	for i := 0; i < 2; i++ {
		x := i
		f := func() int {
			for i := 0; i < 1; i++ {
				x++
			}
			return x
		}
		fns = append(fns, f)
	}
	for _, fn := range fns {
		println(fn())
	}
}

// Output:
// 1
// 2
