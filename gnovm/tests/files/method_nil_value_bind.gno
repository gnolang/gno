// Regression for: `defer i.M()` (or `defer pt.M()`) where the
// interface holds a typed-nil pointer (or pt itself is nil) and M
// has a value receiver must defer the nil-deref panic to call time,
// matching Go. The outer recover catches it after r=1 has run, so
// f() returns 1.

package main

type I interface{ M() }

type T struct {
	x int
}

func (T) M() {}

var pt *T

func f() (r int) {
	defer func() { recover() }()

	var i I = pt
	defer i.M()
	r = 1
	return
}

func main() {
	if got := f(); got != 1 {
		panic(got)
	}
	println("OK")
}

// Output:
// OK
