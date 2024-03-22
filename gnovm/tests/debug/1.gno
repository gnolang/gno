package main

func main() {
	var y, counter int
	var f []func()
	defer func() {
		for _, ff := range f {
			ff()
		}
	}()

	// this is actually a implicit for loop
LABEL_1:
	if counter == 2 {
		return
	}
	x := y
	f = append(f, func() { println(x) })
	y++
	counter++
	goto LABEL_1 // this is the edge condition, break? continue?
}

// Output:
// 1
// 1
