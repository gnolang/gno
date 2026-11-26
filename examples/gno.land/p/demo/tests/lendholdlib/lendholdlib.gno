// Package lendholdlib holds callees for the lending filetests.
package lendholdlib

func Call(f func())                    { f() }
func Pass(f func())                    { Call(f) }
func Keep(f func(), box *[]func())     { *box = append(*box, f) }
func KeepCall(f func(), box *[]func()) { Keep(f, box); (*box)[0]() }
func DeferCall(f func())               { defer f() }
func WrapCall(f func())                { Call(func() { f() }) }
func Hold(g func(), box *[]func())     { Keep(g, box) }

// WrapKeepCall lends a wrapper of f to Hold, which keeps it, then calls it.
func WrapKeepCall(f func(), box *[]func()) {
	Hold(func() { f() }, box)
	(*box)[0]()
}

// PanicIt throws f and calls what it catches while still running.
func PanicIt(f func(), box *[]func()) {
	defer func() {
		if g, ok := recover().(func()); ok {
			*box = append(*box, g)
			g()
		}
	}()
	panic(f)
}

func PanicOut(f func())              { panic(f) }
func CallThenCB(f func(), cb func()) { cb(); f() }
func Recurse(f func(), n int) {
	if n > 0 {
		Recurse(f, n-1)
	}
	f()
}
func Ret(f func() int) int { return f() }
func Branch(f, alt func(), b bool) {
	if b {
		f = alt
	}
	f()
}

type Fn func()

func CallFn(f Fn) { f() }
func (f Fn) Run() { f() }

type Inner struct{}

func (Inner) Do(f func())                  { f() }
func (Inner) Keep(f func(), box *[]func()) { *box = append(*box, f) }

type Outer struct{ Inner }

type Ops struct{ F func() }

func CallOps(o Ops) { o.F() }

type Caller interface{ Do(f func()) }

func Via(c Caller, f func()) { c.Do(f) }
