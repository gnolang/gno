package main

func main() {
	var fns []func() int
	for i := 0; i < 2; i++ {
		for j := 0; j < 2; j++ {
			f := func() int {
				return i + j
			}
			fns = append(fns, f)
		}
	}
	for _, fn := range fns {
		println(fn())
	}
}

// Output:
// 4
// 4
// 4
// 4
