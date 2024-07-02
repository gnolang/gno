package main

func main() {
	var fns []func() int
	for i := 0; i < 3; i++ {
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
