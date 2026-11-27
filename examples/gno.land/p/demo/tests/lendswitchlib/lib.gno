package lendswitchlib

type F func()

// RetSwitch returns the value out of a nested type switch: escapes.
func RetSwitch(f any) any {
	switch g := f.(type) {
	case any:
		switch h := g.(type) {
		case func():
			return h
		}
	}
	return nil
}

// DeferSwitch defers the call: stays in the frame.
func DeferSwitch(f any) {
	switch g := f.(type) {
	case func():
		defer g()
	}
}

// ConvCallee converts to a named func type and calls.
func ConvCallee(f any) {
	switch g := f.(type) {
	case func():
		F(g)()
	}
}

// Into writes the value through a caller pointer: escapes.
func Into(f any, dst *any) {
	switch g := f.(type) {
	case func():
		*dst = g
	}
}

// Loop calls via a goto loop.
func Loop(f any) {
	i := 0
L:
	f.(func())()
	i++
	if i < 3 {
		goto L
	}
}

func boom(x func()) { panic(x) }

// PanicCall gets the value back through panic/recover and calls it in-frame.
func PanicCall(f any) {
	defer func() {
		x := recover()
		x.(func())()
	}()
	switch g := f.(type) {
	case func():
		boom(g)
	}
}

// R lets the lib keep one of two arguments and fire it while the call runs.
type R struct{ keep func() }

func NewR() *R { return &R{} }

func (r *R) Two(a func(), b func()) {
	a()
	r.keep = b
	r.keep()
}

func (r *R) TwoAny(a any, b any) {
	a.(func())()
	r.keep = b.(func())
	r.keep()
}

// Shadow: the switch var is reassigned from another param, then kept.
func Shadow(f any, k any) {
	switch g := f.(type) {
	case func():
		g()
		_ = k
	}
	keep2 := NewR()
	switch g := k.(type) {
	case func():
		keep2.keep = g
		keep2.keep()
	}
}

// DeferAfterPanic runs f in a defer while its frame panics.
func DeferAfterPanic(f func()) {
	defer f()
	panic("x")
}

// RecoverThenCall recovers a panic from a nested call, then calls f again.
func RecoverThenCall(f func()) {
	func() {
		defer func() { recover() }()
		inner(f)
	}()
	f()
}

func inner(f func()) {
	f()
	panic("y")
}

// Rec recurses n deep, calling f at the bottom.
func Rec(n int, f func()) {
	if n == 0 {
		f()
		return
	}
	Rec(n-1, f)
	f()
}
