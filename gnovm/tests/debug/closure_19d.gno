package main

var y, counter int
var f []func()

func main() {
LABEL_1:
	x := y
	if counter == 2 {
		counter = 0
		bar()
		for _, ff := range f { // XXX, why defer on this not work
			ff()
		}
		return
	}
	f = append(f, func() { println(x) })
	y++
	counter++
	goto LABEL_1
}

func bar() {
	println("---bar---")
LABEL_2:
	if counter == 2 {
		println("---end---")
		return
	}
	z := y
	f = append(f, func() { println(z) })
	y++
	counter++
	goto LABEL_2
}

// Output:
// ---bar---
// ---end---
// 0
// 1
// 2
// 3
