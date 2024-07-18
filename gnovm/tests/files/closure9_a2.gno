package main

func main() {
	var fns []func() int
	var x int
	for i := 0; i < 5; i++ {
		x = i
		f := func() int {
			return x
		}
		fns = append(fns, f)
	}
	for _, fn := range fns {
		println(fn())
	}
}

// Go Output:
// 4
// 4
// 4
// 4
// 4

// Output:
// 4
// 4
// 4
// 4
// 4
