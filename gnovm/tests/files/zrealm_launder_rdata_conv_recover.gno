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

func take(launderrvictim.Immutable, bool) {}

// A nil pointer converts; a stored one is refused.
type ImmP *launderrvictim.Immutable

func catch(f func()) (msg string) {
	defer func() {
		if r := recover(); r != nil {
			msg = r.(string)
		}
	}()
	f()
	return "converted"
}

func main(cur realm) {
	println(catch(func() {}))
	println(catch(func() {
		s := struct{ Field string }{"x"}
		_ = launderrvictim.Immutable(s)
	}))
	println(catch(func() { _, _ = ret() }))
	println(catch(func() { launderattack.Convert(laundervictim.GetImm()) }))
	println(catch(func() { take(mk()) }))
	println(catch(func() { defer take(mk()) }))
	println(catch(func() {
		var v launderrvictim.Immutable
		src := map[string]struct{ Field string }{"k": {"x"}}
		v, _ = src["k"]
		_ = v
	}))
	// A refused range writes nothing from its iteration, the key included.
	n, i, k := 0, -1, "unset"
	println(catch(func() {
		var v launderrvictim.Immutable
		for i, v = range []struct{ Field string }{{"x"}} {
			n++
		}
		_ = v
	}), n, i)
	println(catch(func() {
		var v launderrvictim.Immutable
		for k, v = range map[string]struct{ Field string }{"key0": {"x"}} {
			n++
		}
		_ = v
	}), n, k)
	// A range refused on a later iteration keeps the earlier iterations' writes.
	n, i = 0, -1
	println(catch(func() {
		var v ImmP
		for i, v = range []*launderrvictim.Immutable{nil, launderrvictim.GetImm()} {
			n++
		}
		_ = v
	}), n, i)
	println("after")
}

// Output:
// converted
// illegal conversion to external realm type (call a function of the realm that declares the type to make it)

// Error:
// cannot allocate gno.land/r/tests/vm/launderrvictim.Immutable in realm gno.land/r/launderrh_conv_recover

// TypeCheckError:
// gno.land/r/launderrh_conv_recover/zrealm_launder_rdata_conv_recover.gno:15:13: cannot hold a value of gno.land/r/tests/vm/launderrvictim.Immutable outside its realm (use a pointer); gno.land/r/launderrh_conv_recover/zrealm_launder_rdata_conv_recover.gno:17:11: cannot hold a value of gno.land/r/tests/vm/launderrvictim.Immutable outside its realm (use a pointer); gno.land/r/launderrh_conv_recover/zrealm_launder_rdata_conv_recover.gno:36:7: cannot hold a value of gno.land/r/tests/vm/launderrvictim.Immutable outside its realm (use a pointer); gno.land/r/launderrh_conv_recover/zrealm_launder_rdata_conv_recover.gno:38:32: cannot hold a value of gno.land/r/tests/vm/launderrvictim.Immutable outside its realm (use a pointer); gno.land/r/launderrh_conv_recover/zrealm_launder_rdata_conv_recover.gno:40:30: cannot hold a value of gno.land/r/tests/vm/launderrvictim.Immutable outside its realm (use a pointer); gno.land/r/launderrh_conv_recover/zrealm_launder_rdata_conv_recover.gno:41:36: cannot hold a value of gno.land/r/tests/vm/launderrvictim.Immutable outside its realm (use a pointer); gno.land/r/launderrh_conv_recover/zrealm_launder_rdata_conv_recover.gno:43:7: cannot hold a value of gno.land/r/tests/vm/launderrvictim.Immutable outside its realm (use a pointer); gno.land/r/launderrh_conv_recover/zrealm_launder_rdata_conv_recover.gno:45:3: cannot hold a value of gno.land/r/tests/vm/launderrvictim.Immutable outside its realm (use a pointer); gno.land/r/launderrh_conv_recover/zrealm_launder_rdata_conv_recover.gno:46:7: cannot hold a value of gno.land/r/tests/vm/launderrvictim.Immutable outside its realm (use a pointer); gno.land/r/launderrh_conv_recover/zrealm_launder_rdata_conv_recover.gno:51:7: cannot hold a value of gno.land/r/tests/vm/launderrvictim.Immutable outside its realm (use a pointer); gno.land/r/launderrh_conv_recover/zrealm_launder_rdata_conv_recover.gno:52:10: cannot hold a value of gno.land/r/tests/vm/launderrvictim.Immutable outside its realm (use a pointer); gno.land/r/launderrh_conv_recover/zrealm_launder_rdata_conv_recover.gno:55:7: cannot hold a value of gno.land/r/tests/vm/launderrvictim.Immutable outside its realm (use a pointer); gno.land/r/launderrh_conv_recover/zrealm_launder_rdata_conv_recover.gno:58:7: cannot hold a value of gno.land/r/tests/vm/launderrvictim.Immutable outside its realm (use a pointer); gno.land/r/launderrh_conv_recover/zrealm_launder_rdata_conv_recover.gno:59:10: cannot hold a value of gno.land/r/tests/vm/launderrvictim.Immutable outside its realm (use a pointer); gno.land/r/launderrh_conv_recover/zrealm_launder_rdata_conv_recover.gno:62:7: cannot hold a value of gno.land/r/tests/vm/launderrvictim.Immutable outside its realm (use a pointer)
