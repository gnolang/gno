package main

// Init RHS references that must still be renamed: an enclosing loop's own
// variable. The inner init `i := i * 10` reads the outer loop's i.loopvar.
func nested() {
	var fns []func() int
	for i := 0; i < 2; i++ {
		for i := i * 10; i%10 < 2; i++ {
			fns = append(fns, func() int { return i })
		}
	}
	for _, fn := range fns {
		println(fn())
	}
}

// A func literal in the init RHS captures the outer x, not the loop's x.
func initFuncLit() {
	x := 7
	for x := func() int { return x + 1 }(); x < 10; x++ {
		println(x)
	}
	println(x)
}

// Multi-name init: the later RHS reads the outer i, not the new one.
func multiInit() {
	i := 100
	for i, j := 0, i; i < 2; i++ {
		println(i, j)
	}
	println(i)
}

// A plain redefinition between two same-named loops: the inner init RHS
// reads the plain (non-loopvar) x, so it is not renamed.
func shadowBetween() {
	for x := 0; x < 1; x++ {
		x := 5
		for x := x; x < 7; x++ {
			println(x)
		}
	}
}

func main() {
	nested()
	initFuncLit()
	multiInit()
	shadowBetween()
}

// Output:
// 0
// 1
// 10
// 11
// 8
// 9
// 7
// 0 100
// 1 100
// 100
// 5
// 6
