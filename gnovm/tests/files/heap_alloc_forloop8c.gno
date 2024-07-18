package main

func main() {
	var fns []func() int

	for i := 0; i < 2; i++ {
		x := i
		y := 1
		f := func() int {
			switch y {
			case 1:
				x += 1
			default:
				x += 0
			}
			return x
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
