package main

// the vp of the arg of the outer closure should be (1,0)
// the vp of the x of inner closure should be (4,0)
func main() {
	var fns []func() int

	for i := 0; i < 2; i++ {
		x := i
		f := func() int {
			if true {
				if true {
					return x
				}
			}
			return 0
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
