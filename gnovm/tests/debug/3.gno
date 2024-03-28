package main

func main() {
	var fns []func() int

	for i := 0; i < 5; i++ {
		x := i
		f := func() int {
			return x
		}
		fns = append(fns, f)
		x += 1
	}
	for _, fn := range fns {
		println(fn())
	}
}

// Output:
// 1
// 2
// 3
// 4
// 5
