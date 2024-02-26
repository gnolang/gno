package main

func main() {
	var y, counter int

	println(&y)

	var f []func() *int

	defer func() {
		for _, ff := range f {
			println(ff())
			println(*ff())
		}
	}()

	// this is actually a implicit for loop
LABEL_1:
	if counter == 2 {
		return
	}
	f = append(f, func() *int {
		//x := &y
		return &y
	})
	y++
	counter++
	goto LABEL_1 // this is the edge condition, break? continue?
}

// Output:
// &0x14000283f80.(*int)
// &0x14000283f80.(*int)
// 1
// &0x14000283f80.(*int)
// 2
