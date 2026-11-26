// PKGPATH: gno.land/r/tests/vm/lendaliasstored
package lendaliasstored

import lib "gno.land/p/demo/tests/t76lib"

// A crossing literal kept in the realm and run in a later crossing call: one
// that captured an alias must do what its by-name twin does.

var (
	n       int
	xs      = []int{1, 2, 3}
	byName  func(cur realm)
	byAlias func(cur realm)
)

func keep(cur realm) {
	byName = func(cur realm) { lib.Each(xs, func(x int) { n += x }) }
	each := lib.Each
	byAlias = func(cur realm) { each(xs, func(x int) { n += x }) }
}

func show(s string, err any) {
	println(s, n, err)
	n = 0
}

func main(cur realm) {
	keep(cross(cur))
	show("stored crossing literal, name ", try(func() { byName(cross(cur)) }))
	show("stored crossing literal, alias", try(func() { byAlias(cross(cur)) }))
}

func try(fn func()) (r any) {
	defer func() { r = recover() }()
	fn()
	return nil
}

// Output:
// stored crossing literal, name  6 undefined
// stored crossing literal, alias 6 undefined
