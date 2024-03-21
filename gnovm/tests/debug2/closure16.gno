package main

func main() {
	var fns []func() int
	s := []int{1, 2, 3}
	for i, _ := range s {
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

// Output:
// 0
// 1
// 2
