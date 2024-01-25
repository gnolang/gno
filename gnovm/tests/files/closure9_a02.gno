package main

func main() {
	var fns []func() int
	for i := 0; i < 5; i++ {
		x := i
		f := func() int {
			return x
		}
		fns = append(fns, f)
	}
	println(fns[3]())
	println(fns[2]())
	println(fns[0]())
}

// Output:
// 3
// 2
// 0
