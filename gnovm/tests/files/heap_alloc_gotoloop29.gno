package main

func main() {
	counter1 := 0

	y := 0

	var fs []func()

	defer func() {
		for _, ff := range fs {
			ff()
		}
	}()

LOOP:
	y
	if counter1 < 2 {
		fs = append(fs, func() { println(y) })
		y++
		counter1++
		goto LOOP
	} else {
		return
	}
}

// Output:
// 2
// 2
