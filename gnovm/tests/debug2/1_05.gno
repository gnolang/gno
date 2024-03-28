package main

func main() {
	var y, counter int
	var f []func()
	defer func() {
		for _, ff := range f {
			ff()
		}
	}()

	for i := 0; i < 2; i++ {
		counter = 0
	LABEL_1:
		if counter == 2 {
			continue
		}
		x := y
		f = append(f, func() { println(x) })
		y++
		counter++
		goto LABEL_1
	}
}

// Output:
// 0
// 1
// 2
// 3
