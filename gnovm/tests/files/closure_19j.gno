package main

func main() {
	var y, counter int
	var f []func()
	defer func() {
		for _, ff := range f {
			ff()
		}
	}()

	f1 := func() {
	LABEL_1:
		if counter == 2 {
			return
		}
		x := y
		f = append(f, func() { println(x) })
		y++
		counter++
		goto LABEL_1
	}

	f1()
}

// Preprocessed:

// Output:
// undefined
// undefined
