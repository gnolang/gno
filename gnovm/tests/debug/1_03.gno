package main

func main() {
	var y, counter int
	var f []func() func() int
	defer func() {
		for _, ff := range f {
			println(ff()())
		}
	}()

LABEL_1:
	if counter == 2 {
		return
	}
	x := y
	f = append(f, func() func() int {
		return func() int { return x }
	})
	y++
	counter++
	goto LABEL_1
}

// Output:
// 0
// 1
