package main

func main() {
	var fns []func() int
	for i := 0; i < 5; i++ {
		f := func() int {
			return i
		}
		fns = append(fns, f)
	}
	for _, fn := range fns {
		println(fn())
	}
}

// go 1.22 loop var is not supported for now.

// Output:
// 5
// 5
// 5
// 5
// 5
