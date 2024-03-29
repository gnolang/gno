package main

func main() {
	var y, counter int
	var f []func()
	defer func() {
		for _, ff := range f {
			ff()
		}
	}()

	{
	LABEL_1:
		x := y
		if counter == 2 {
			counter = 0
			goto LABEL_2
		}
		f = append(f, func() { println(x) })
		y++
		counter++
		goto LABEL_1
	}

LABEL_2:
	if counter == 2 {
		return
	}
	z := y
	f = append(f, func() { println(z) })
	y++
	counter++
	goto LABEL_2
}

// Output:
// 0
// 1
// 2
// 3
