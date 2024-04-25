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
			println(i*100 + v)
		})
		y++
	}
}

// Output:
// 1
// 102
// 203
