package main

// this is a fix, intuitive, and same with Go
func main() {
	var fns []func() int
	//defer func() {
	//	for _, fn := range fns {
	//		println(fn())
	//	}
	//}()

	for i := 0; i < 5; i++ {
		//defer func() {
		//	for _, fn := range fns {
		//		println(fn())
		//	}
		//}()
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
// 3
// 4
