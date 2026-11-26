// Package t76victim calls a caller's function value captured by a literal
// converted to a named func type, next to a same-named local declared later.
package t76victim

import lib "gno.land/p/demo/tests/t76lib"

var N int

func bump() { N += 1 }

// ExName is a non-crossing writer a caller can pass back by value.
func ExName() { bump() }

// Do runs one case inside a crossing call and reports its panic.
func Do(cur realm, which string, f func(), g func(int)) (r string) {
	defer func() {
		if e := recover(); e != nil {
			r = "REFUSED"
			if s, ok := e.(string); ok {
				r = s
			}
		}
	}()
	switch which {
	case "CapCtl":
		run(func() {
			f()
		})
	case "CapParam":
		run(func() {
			f()
			f := noop
			f()
		})
	case "CapLocal":
		h := f
		run(func() {
			h()
			h := noop
			h()
		})
	case "CapHelper":
		run(func() {
			f()
			f := func() {}
			f()
		})
	case "CapG":
		run(func() {
			g(5)
			g := lib.Call
			g(nil)
		})
	case "CbLibName":
		lib.Each([]int{1, 2, 3}, g)
	default:
		panic("no case " + which)
	}
	return "ok"
}

// Ptr hands out a pointer to N: a view for another realm.
func Ptr() *int { return &N }

func Get() int { return N }

func Reset(cur realm) { N = 0 }

type Fn func()

func run(f Fn) { f() }
func noop()    {}
