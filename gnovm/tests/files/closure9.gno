package main

// this is a fix, intuitive, and same with Go
func main() {
	var fns []func() int
	for i := 0; i < 5; i++ {
		x := i
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
// 0
// 1
// 2
// 3
// 4

// Output:
// 0
// 1
// 2
// 3
// 4
