package main

func main() {
	var y int
	var f []func()

	defer func() {
		for _, ff := range f {
			ff()
		}
	}()

	// on the re-preprocess process,
	// loop var and loop extern no need
	// to adjust depth [a,i], others all need[f, y]
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
