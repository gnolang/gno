// A STORED method value over a shadowed shape: FS's unexported main-package
// FIELD sec must not be matched when the bind formed in secx is called later
// (identity is package-qualified; the bind carries MethodPkg). Also pins the
// deferred-call resolution path (defer g()) and value-operand snapshot with
// repeated calls. Verified against real Go.
package main

import "filetests/extern/secx"

type FS struct {
	sec int
	secx.SecImpl
}

func main() {
	f := FS{}
	f.sec = 42
	g := secx.MakeGetter(f)
	println("B:", g(), g())

	func() {
		g2 := secx.MakeGetter(f)
		defer g2()
		println("B2: pre-defer")
	}()
	println("B2: done")

	c := secx.Cnt{N: 1}
	gg := secx.MakeGetter(c)
	c.N = 99
	println("G:", gg(), gg())
}

// Output:
// B: 7 7
// B2: pre-defer
// B2: done
// G: 1 1
