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
FOO_LABEL:
	if counter == 5 {
		//for _, ff := range f {
		//	ff()
		//}
		return
	}
	x := y
	f = append(f, func() { println(x) })
	x += 1
	y++
	counter++
	goto FOO_LABEL // this is the edge condition, break? continue?
}
