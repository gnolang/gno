// PKGPATH: gno.land/r/launderrh_conv_recover
package launderrh_conv_recover

import (
	"gno.land/p/demo/tests/launderattack/v0"
	"gno.land/r/tests/vm/launderrvictim"
	"gno.land/r/tests/vm/laundervictim"
)

// A refused conversion is a panic a realm can recover, for `T(x)` and for an
// implicit source alike.

func mk() (struct{ Field string }, bool) { return struct{ Field string }{"x"}, true }

func ret() (launderrvictim.Immutable, bool) { return mk() }

func catch(f func()) (msg string) {
	defer func() { msg = recover().(string) }()
	f()
	return "converted"
}

func main(cur realm) {
	println(catch(func() {
		s := struct{ Field string }{"x"}
		_ = launderrvictim.Immutable(s)
	}))
	println(catch(func() { _, _ = ret() }))
	println(catch(func() { launderattack.Convert(laundervictim.GetImm()) }))
	println("after")
}

// Output:
// illegal conversion to external realm type
// illegal conversion to external realm type
// illegal conversion of readonly or externally stored value
// after
