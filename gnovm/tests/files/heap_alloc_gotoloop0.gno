package main

func main() {
	var counter int
	var f []func()
	defer func() {
		for _, ff := range f {
			ff()
		}
	}()

	// this is actually an implicit for loop
LABEL_1:
	if counter == 2 {
		return
	}
	x := counter
	f = append(f, func() { println(x) })
	counter++
	goto LABEL_1
}

// Output:
// 0
// 1
