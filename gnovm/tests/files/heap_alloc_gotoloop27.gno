package main

func main() {
	var y int
	var f []func()
	defer func() {
		for _, ff := range f {
			ff()
		}
	}()

	for i := 0; i < 2; i++ {
		for j := 0; j < 2; j++ { // every inner loop will create a new block too, see 1_05
			x := y
			f = append(f, func() { println(x) })
			y++
		}
	}
}

// Output:
// 0
// 1
// 2
// 3
