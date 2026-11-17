// Package libtable holds functions reachable through package-level variables:
// a closure, a table of functions, and a dispatcher over its own table.
package libtable

type Box struct{ N int }

func Run(f func()) { f() }

var Apply = func(f func()) { f() }

var Table = map[string]func(func()){"run": Run}

func Each(bs []*Box, f func(*Box)) {
	for _, b := range bs {
		f(b)
	}
}

func Inc(b *Box) { b.N++ }

var Handlers = map[string]func(*Box){"inc": Inc}

// Dispatch reaches Inc through this package's own table.
func Dispatch(k string, b *Box) { Handlers[k](b) }

func Panics() { panic("boom") }
