package main

func main() {
	var fns []func() int
	for i := 0; i < 5; i++ {
		var x int
		f := func() int {
			return x
		}
		x = i
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
