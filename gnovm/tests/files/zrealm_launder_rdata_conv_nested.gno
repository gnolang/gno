// PKGPATH: gno.land/r/launderrh_conv_nested
package launderrh_conv_nested

// `return f()`, `g(f())`, comma-ok and range refuse to convert a value another
// realm stores, read at the top of that realm's value or below it: a row of a
// stored [][]string, a slice field of a stored struct or two structs down, a
// map value.

import "gno.land/r/tests/vm/launderrdeep"

type L []string
type LL [][]string

func row() ([]string, bool)    { return launderrdeep.Grid()[0], true }
func grid() ([][]string, bool) { return launderrdeep.Grid(), true }
func items() ([]string, bool)  { return launderrdeep.GetBox().Items, true }

func deep() ([]string, bool)    { return launderrdeep.GetOuter().Mid.Items, true }
func deepPtr() ([]string, bool) { return launderrdeep.GetOuterPtr().Mid.Items, true }

func retRow() (L, bool)     { return row() }
func retGrid() (LL, bool)   { return grid() }
func retDeep() (L, bool)    { return deep() }
func retDeepPtr() (L, bool) { return deepPtr() }

func lvl3() ([]string, bool) { return row() }
func lvl2() ([]string, bool) { return lvl3() }
func lvl1() (L, bool)        { return lvl2() }

func take(l L, ok bool) {}

func try(label string, fn func()) {
	defer func() {
		r := recover()
		println(label, "->", r)
	}()
	fn()
}

func main(cur realm) {
	try("return f(), row of [][]string", func() { retRow() })
	try("return f(), whole [][]string", func() { retGrid() })
	try("return f(), three calls deep", func() { lvl1() })
	try("return f(), two structs down, by value", func() { retDeep() })
	try("return f(), two structs down, by pointer", func() { retDeepPtr() })
	try("g(f()), slice field of a struct", func() { take(items()) })
	try("comma-ok, map value [][]string", func() {
		var v LL
		var ok bool
		v, ok = launderrdeep.Table()["k"]
		_, _ = v, ok
	})
	try("comma-ok, assertion on a row", func() {
		var x any = launderrdeep.Grid()[1]
		var v L
		var ok bool
		v, ok = x.([]string)
		_, _ = v, ok
	})
	try("range, rows of [][]string", func() {
		var r L
		for _, r = range launderrdeep.Grid() {
			_ = r
		}
	})
	try("range, map of [][]string", func() {
		var k string
		var v LL
		for k, v = range launderrdeep.Table() {
			_, _ = k, v
		}
	})
}

// Output:
// return f(), row of [][]string -> illegal conversion of readonly or externally stored value (convert it inside a crossing function of the realm that stores it)
// return f(), whole [][]string -> illegal conversion of readonly or externally stored value (convert it inside a crossing function of the realm that stores it)
// return f(), three calls deep -> illegal conversion of readonly or externally stored value (convert it inside a crossing function of the realm that stores it)
// return f(), two structs down, by value -> illegal conversion of readonly or externally stored value (convert it inside a crossing function of the realm that stores it)
// return f(), two structs down, by pointer -> illegal conversion of readonly or externally stored value (convert it inside a crossing function of the realm that stores it)
// g(f()), slice field of a struct -> illegal conversion of readonly or externally stored value (convert it inside a crossing function of the realm that stores it)
// comma-ok, map value [][]string -> illegal conversion of readonly or externally stored value (convert it inside a crossing function of the realm that stores it)
// comma-ok, assertion on a row -> illegal conversion of readonly or externally stored value (convert it inside a crossing function of the realm that stores it)
// range, rows of [][]string -> illegal conversion of readonly or externally stored value (convert it inside a crossing function of the realm that stores it)
// range, map of [][]string -> illegal conversion of readonly or externally stored value (convert it inside a crossing function of the realm that stores it)
