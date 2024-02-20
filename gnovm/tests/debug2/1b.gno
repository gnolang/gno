package main

// invalid goto
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
	if counter == 5 {
		goto LABEL_2
	}
	x := y
	f = append(f, func() { println(x) })
	y++
	counter++
	goto LABEL_1 // this is the edge condition, break? continue?

LABEL_2:
	println("end")
}

// Output:
