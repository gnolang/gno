package main

func main() {
	var y, counter int
	var f []func() (int, func() int)
	defer func() {
		for _, ff := range f {
			n, f := ff()
			println(n + f())
		}
	}()

LABEL_1:
	if counter == 2 {
		return
	}
	x := y
	z := y
	f = append(f, func() (int, func() int) {
		return z, func() int { return x }
	})
	y++
	counter++
	goto LABEL_1
}

// Output:
// 0
// 2
