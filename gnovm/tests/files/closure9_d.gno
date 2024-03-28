package main

func main() {
	var fns []func() int
	for i := 0; i < 5; i++ {
		x := i
		f := func() int {
			x := 5
			return x
		}
		println(x)
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
// 3
// 4
// 5
// 5
// 5
// 5
// 5
