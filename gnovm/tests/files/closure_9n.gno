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
		a := i
		f = append(f, func() {
			println(a)
		})
		y++
	}
}

// Output:
// 0
// 1
// 2
