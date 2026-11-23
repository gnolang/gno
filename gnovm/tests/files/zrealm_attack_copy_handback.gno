// PKGPATH: gno.land/r/attackcopyhandback
// R3 copy-and-hand-back: a copy of another realm's value is not the owner's
// value, so the owner's crossing function must not trust it back.
package attackcopyhandback

import "gno.land/r/tests/vm/fvx"

func main(cur realm) {
	p := fvx.New(1)
	c := *p
	c.N = 999
	println(fvx.Read(cross(cur), &c))
}

// Error:
// cannot copy gno.land/r/tests/vm/fvx.T outside realm gno.land/r/tests/vm/fvx (use a pointer)

// TypeCheckError:
// gno.land/r/attackcopyhandback/zrealm_attack_copy_handback.gno:10:2: cannot hold a value of gno.land/r/tests/vm/fvx.T outside its realm (use a pointer); gno.land/r/attackcopyhandback/zrealm_attack_copy_handback.gno:10:7: cannot hold a value of gno.land/r/tests/vm/fvx.T outside its realm (use a pointer)
