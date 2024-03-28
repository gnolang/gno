package main

// this is a fix, intuitive, and same with Go
func main() {
	var fns []func() int

	for i := 0; i < 5; i++ {
		x := i
		if i <= 2 {
			continue
		}

		f := func() int {
			return x
		}
		fns = append(fns, f)
		x += 1
	}
	for _, fn := range fns {
		println(fn())
	}
}

// Output:
// 4
// 5
