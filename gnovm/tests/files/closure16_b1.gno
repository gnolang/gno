package main

func main() {
	var fns []func() int
	s := "hello"
	for i, _ := range s {
		f := func() int {
			return i
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
// 3
// 4
