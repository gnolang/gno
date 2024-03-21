package main

func main() {
	var y int
	var f []func()

	defer func() {
		for _, ff := range f {
			ff()
		}
	}()

	s := []int{1, 2, 3}
	for i, v := range s {
		f = append(f, func() {
			x := i // i is loop extern, see 1a3
			println(x*100 + v)
		})
		y++
	}
}

// Output:
// 203
// 203
// 203
