package main

func main() {
	var fns []func() int
	y := 0
	m := map[string]int{"a": 1, "b": 2}
	for _, v := range m {
		x := v
		f := func() int {
			switch y {
			case 0:
				if true {
					return x
				}
			default:
				return 0
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
// 1
// 2
