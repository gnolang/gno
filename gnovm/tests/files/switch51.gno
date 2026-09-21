package main

// Heap-captured clause names on both sides of a fallthrough keep their own
// slots: the closure over x must not see the value later written to y.
func main() {
	fs := []func(){}
	switch 1 {
	case 1:
		x := 1
		fs = append(fs, func() { println("x", x) })
		fallthrough
	case 2:
		y := 2
		fs = append(fs, func() { println("y", y) })
	}
	for _, f := range fs {
		f()
	}
}

// Output:
// x 1
// y 2
