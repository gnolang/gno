// PKGPATH: gno.land/r/tests/vm/lendaliasshadowown
package lendaliasshadowown

// A value captured by a literal converted to a named func type stays a value
// when the literal later declares a same-named alias or helper.

import (
	lib "gno.land/p/demo/tests/t76lib"
)

type Fn func()

var (
	n      int
	xs     = []int{1, 2, 3}
	stored func([]int, func(int))
)

func init() {
	stored = func(xs []int, f func(int)) { n += 100 }
}

func show(s string, err any) {
	if e, ok := err.(string); ok && len(e) > 70 {
		err = e[30:70]
	}
	println(s, n, err)
	n = 0
}

func run(f Fn) { f() }

func main(cur realm) {
	show("x0 control: stored by value in converted literal", try(func() {
		each := stored
		run(func() {
			each(xs, func(x int) { n += x })
		})
	}))
	show("x1 captured outer value called before a same-named alias", try(func() {
		each := stored
		run(func() {
			each(xs, func(x int) { n += x })
			each := lib.Each
			each(nil, func(x int) {})
		})
	}))
	show("x2 captured outer value called before a same-named helper", try(func() {
		each := stored
		run(func() {
			each(xs, func(x int) { n += x })
			each := func(xs []int, f func(int)) {}
			each(nil, nil)
		})
	}))
	show("x1t twin: plain literal, not converted", try(func() {
		each := stored
		func() {
			each(xs, func(x int) { n += x })
			each := lib.Each
			each(nil, func(x int) {})
		}()
	}))
}

func try(fn func()) (r any) {
	defer func() { r = recover() }()
	fn()
	return nil
}

// Output:
// x0 control: stored by value in converted literal 0 hadowown's code is read-only here: a fun
// x1 captured outer value called before a same-named alias 0 hadowown's code is read-only here: a fun
// x2 captured outer value called before a same-named helper 0 hadowown's code is read-only here: a fun
// x1t twin: plain literal, not converted 0 hadowown's code is read-only here: a fun
