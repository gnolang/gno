package main

// Same shadowing init as loopvar_shadow_init_capture_1, but the closures
// are inside the body and capture the loop's own x: each iteration must
// still get its own copy (go1.22 semantics), seeded from the outer x.
func main() {
	var fns []func() int
	x := 7
	for x, i := x, 0; i < 3; i++ {
		x += i
		fns = append(fns, func() int { return x })
	}
	for _, fn := range fns {
		println(fn())
	}
	println(x)
}

// Output:
// 7
// 8
// 10
// 7
