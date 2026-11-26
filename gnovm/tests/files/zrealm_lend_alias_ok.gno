// PKGPATH: gno.land/r/tests/vm/lendaliasok
package lendaliasok

import (
	"sort"

	lib "gno.land/p/demo/tests/t76lib"
)

// A local declared once from a declared function and only ever called is that
// function by name: each "alias" line must equal its "name" twin.

var (
	n  int
	xs = []int{1, 2, 3}
	ys []int
)

func own(xs []int, f func(int)) {
	for _, x := range xs {
		f(x)
	}
}

func show(s string, err any) {
	println(s, n, err)
	n = 0
}

func main(cur realm) {
	show("01 /p/ lent literal, name ", try(func() { lib.Each(xs, func(x int) { n += x }) }))
	show("01 /p/ lent literal, alias", try(func() { each := lib.Each; each(xs, func(x int) { n += x }) }))

	show("02 own function, name ", try(func() { own(xs, func(x int) { n += x }) }))
	show("02 own function, alias", try(func() { f := own; f(xs, func(x int) { n += x }) }))

	show("03 var form, name ", try(func() { lib.Each(xs, func(x int) { n += x }) }))
	show("03 var form, alias", try(func() { var each = lib.Each; each(xs, func(x int) { n += x }) }))

	show("04 inside literal called in place, name ", try(func() {
		func() { lib.Each(xs, func(x int) { n += x }) }()
	}))
	show("04 inside literal called in place, alias", try(func() {
		each := lib.Each
		func() { each(xs, func(x int) { n += x }) }()
	}))

	show("05 inner block, name ", try(func() {
		if len(xs) > 0 {
			lib.Each(xs, func(x int) { n += x })
		}
	}))
	show("05 inner block, alias", try(func() {
		if len(xs) > 0 {
			each := lib.Each
			each(xs, func(x int) { n += x })
		}
	}))

	show("06 two aliases, name ", try(func() {
		lib.Each(xs, func(x int) { n += x })
		lib.Each(xs, func(x int) { n += 10 * x })
	}))
	show("06 two aliases, alias", try(func() {
		a := lib.Each
		b := lib.Each
		a(xs, func(x int) { n += x })
		b(xs, func(x int) { n += 10 * x })
	}))

	show("07 loop, name ", try(func() {
		for i := 0; i < 3; i++ {
			lib.Each(xs, func(x int) { n += x })
		}
	}))
	show("07 loop, alias", try(func() {
		each := lib.Each
		for i := 0; i < 3; i++ {
			each(xs, func(x int) { n += x })
		}
	}))

	show("08 inside a local helper, name ", try(func() {
		h := func() { lib.Each(xs, func(x int) { n += x }) }
		h()
	}))
	show("08 inside a local helper, alias", try(func() {
		each := lib.Each
		h := func() { each(xs, func(x int) { n += x }) }
		h()
	}))

	show("09 stdlib on own slice, name ", try(func() { ys = []int{3, 1, 2}; sort.Ints(ys); n = ys[0] }))
	show("09 stdlib on own slice, alias", try(func() { ys = []int{3, 1, 2}; f := sort.Ints; f(ys); n = ys[0] }))
}

func try(fn func()) (r any) {
	defer func() { r = recover() }()
	fn()
	return nil
}

// Output:
// 01 /p/ lent literal, name  6 undefined
// 01 /p/ lent literal, alias 6 undefined
// 02 own function, name  6 undefined
// 02 own function, alias 6 undefined
// 03 var form, name  6 undefined
// 03 var form, alias 6 undefined
// 04 inside literal called in place, name  6 undefined
// 04 inside literal called in place, alias 6 undefined
// 05 inner block, name  6 undefined
// 05 inner block, alias 6 undefined
// 06 two aliases, name  66 undefined
// 06 two aliases, alias 66 undefined
// 07 loop, name  18 undefined
// 07 loop, alias 18 undefined
// 08 inside a local helper, name  6 undefined
// 08 inside a local helper, alias 6 undefined
// 09 stdlib on own slice, name  1 undefined
// 09 stdlib on own slice, alias 1 undefined
