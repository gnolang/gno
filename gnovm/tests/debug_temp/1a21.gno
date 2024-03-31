package main

func main() {
	var y int
	var f []func()

	defer func() {
		for _, ff := range f {
			ff()
		}
	}()

	i := 0
	for i = 0; i < 3; i++ {
		f = append(f, func() {
			x := i // i is loop extern, see 1a3
			println(x)
		})
		y++
	}
}

// Output:
// 3
// 3
// 3
