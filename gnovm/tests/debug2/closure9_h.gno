package main

func main() {
	var fns []func() int
	for i := 0; i < 2; i++ {
		x := i
		for j := 0; j < 2; j++ {
			y := j
			f := func() int {
				return x + y
			}
			fns = append(fns, f)
		}
	}
	for _, fn := range fns {
		println(fn())
	}
}

// Go Output:
// 0
// 1
// 1
// 2

// Output:
// 0
// 1
// 1
// 2
