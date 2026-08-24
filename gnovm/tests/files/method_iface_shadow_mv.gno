// A STORED method value over a shadowed shape: FS's unexported main-package
// FIELD sec must not be matched when the bind formed in secx is called later
// (identity is package-qualified; the bind carries MethodPkg). Also pins the
// deferred-call resolution path — a recording receiver shows the METHOD ran
// (not the field), exactly once, only after the enclosing func returns — and
// value-operand snapshot with repeated calls. Verified against real Go.
package main

import "filetests/extern/secx"

type FS struct {
	sec int
	secx.SecImpl
}

// FR pairs the shadowing unexported field with an embedded recording impl.
type FR struct {
	sec int
	*secx.RecImpl
}

func main() {
	f := FS{}
	f.sec = 42
	g := secx.MakeGetter(f)
	println("B:", g(), g())

	r := &secx.RecImpl{}
	func() {
		g2 := secx.MakeGetter(FR{RecImpl: r})
		defer g2()
		println("B2: pre-defer, ran =", r.Ran)
	}()
	println("B2: done, ran =", r.Ran)

	c := secx.Cnt{N: 1}
	gg := secx.MakeGetter(c)
	c.N = 99
	println("G:", gg(), gg())
}

// Output:
// B: 7 7
// B2: pre-defer, ran = 0
// B2: done, ran = 1
// G: 1 1
