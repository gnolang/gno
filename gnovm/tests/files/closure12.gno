package main

func main() {
	var fns []func() int

	for i := 0; i < 2; i++ {
		x := i
		f := func() int {
			{
				return x
			}
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
