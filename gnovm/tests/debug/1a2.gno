package main

func main() {
	var y int
	var f []func()

	defer func() {
		for _, ff := range f {
			ff()
		}
	}()

	for i := 0; i < 3; i++ {
		//{
		//   i := i
		f = append(f, func() {
			x := i // i is loop extern, see 1a3
			println(x)
		})
		y++
		//}
	}
}

// Output:
// 0
// 1
// 2
