package main

// explain:
// in the 1st loop, fns has a fn capture x with value 0,
// and a defer func is placed with refs to fns(it's not captured in face),
// in the 2nd loop, fn captured x with value 1, and fn appended to fns.
// so here, the first defer will print, 0 , 1.
// still in the 2nd loop, a second defer is placed, with fn[0] and fn[1]
// both capture the x in context, whose value is 1, so the second defer
// will print 1, 1
func main() {
	var fns []func() int

	println("start for loop")
	for i := 0; i < 2; i++ {
		defer func() { // possible resource leak?
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
// 1
// 1
