package main

func main() {
	var fns []func() int

	println("start for loop")
	for i := 0; i < 2; i++ {
		defer func() {
			println("defer")
			for _, fn := range fns {
				println(fn())
			}
		}()

		x := i
		f := func() int {
			return x
		}

		fns = append(fns, f)
	}
	println("end for loop")
}

// Output:
// start for loop
// end for loop
// defer
// 0
// 1
// defer
// 0
// 1
