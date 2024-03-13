package main

func main() {
	var fns []func() int
	for i := 0; i < 2; i++ {
		x := i
		y := 0
		f := func() int {
			x += y
			x += 1
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
