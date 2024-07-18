package main

func main() {
	var fns []func() int
	var x int
	for i := 0; i < 2; i++ {
		x = i
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
// 1
// 2
// 1
// 2

// Output:
// 1
// 2
// 1
// 2
