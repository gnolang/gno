package main

func main() {
	var fns []func() int

	for i := 0; i < 2; i++ {
		x := i
		s := []int{1, 2}

		f := func() int {
			for _, v := range s {
				x += v
			}
			return x
		}
		fns = append(fns, f)
	}
	for _, fn := range fns {
		println(fn())
	}
}

// Output:
// 3
// 4
