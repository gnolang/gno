package main

func main() {
	var fns []func() int
	m := map[string]int{"a": 1, "b": 2}
	for _, v := range m {
		f := func() int {
			return v
		}
		fns = append(fns, f)
	}
	for _, fn := range fns {
		println(fn())
	}
}

// Output:
// 1
// 2
