// PKGPATH: gno.land/r/tests/vm/lendchaincrossing
package lendchaincrossing

var count int

func each2(xs []int, f func(int)) {
	for _, x := range xs {
		f(x)
	}
}

func each(xs []int, f func(int)) { each2(xs, f) }

func hop(cur realm, xs []int, f func(int)) { each2(xs, f) }

func mid0(xs []int, f func(int), rlm realm) { each(xs, f) }

func mid(xs []int, f func(int), rlm realm) { hop(cross(rlm), xs, f) }

// A lent literal passed on through a new crossing call of the same realm,
// its lending frame still live, is refused there (the panic crosses a realm
// boundary, so it ends the run).
func main(cur realm) {
	xs := []int{1, 2, 3}
	r := try(func() { each(xs, func(x int) { count += x }) })
	println("x01 control: two hops in the lending crossing", count, r)
	mid0(xs, func(x int) { count += x }, cur)
	println("x02 control: a realm value passed, no new crossing", count)
	mid(xs, func(x int) { count += x }, cur)
	println("x03 not reached", count)
}

func try(fn func()) (r any) {
	defer func() { r = recover() }()
	fn()
	return nil
}

// Output:
// x01 control: two hops in the lending crossing 6 undefined
// x02 control: a realm value passed, no new crossing 12

// Error:
// gno.land/r/tests/vm/lendchaincrossing's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/lendchaincrossing) package{})).count
