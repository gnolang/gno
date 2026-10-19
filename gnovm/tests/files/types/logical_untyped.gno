package main

import "fmt"

type B bool

func (b B) Name() string {
	if b {
		return "B(true)"
	}
	return "B(false)"
}

func g(b bool) bool { return b }

func main() {
	x, y := 1, 2
	// A named bool context reaches the comparison operands.
	var b B = x < y && true
	fmt.Printf("%T %s\n", b, b.Name())
	var e B = (x < y && x < y) || x > y
	fmt.Printf("%T %s\n", e, e.Name())
	// Without a context the result is plain bool, also when nested.
	c := x < y && true
	fmt.Printf("%T %v\n", c, c)
	d := (x < y && x < y) || x > y
	fmt.Printf("%T %v %v\n", d, d, g(d))
	f := x < y && (x < y || x > y)
	var h bool = f
	fmt.Printf("%T %v\n", h, h)
	for ld := (x < y && x < y) || x > y; ld; {
		fmt.Printf("%T %v\n", ld, ld)
		break
	}
	var i interface{} = x < y && x < y
	fmt.Printf("%T %v\n", i, i)
}

// Output:
// main.B B(true)
// main.B B(true)
// bool true
// bool true true
// bool true
// bool true
// bool true
