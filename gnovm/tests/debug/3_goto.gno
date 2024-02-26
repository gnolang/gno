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
		x += 1
		if i == 2 {
			goto END
		}
	}
END:
	for _, fn := range fns {
		println(fn())
	}
}

// Output:
// 1
// 2
// 3
