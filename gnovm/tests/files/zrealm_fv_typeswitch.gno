// PKGPATH: gno.land/r/fvtypeswitch
package fvtypeswitch

import "gno.land/r/tests/vm/fvx"

func main(cur realm) {
	switch v := fvx.Fail().(type) {
	case fvx.E:
		println(v.Msg)
	}
}

// Error:
// cannot copy gno.land/r/tests/vm/fvx.E outside realm gno.land/r/tests/vm/fvx (use a pointer)

// TypeCheckError:
// gno.land/r/fvtypeswitch/zrealm_fv_typeswitch.gno:8:2: cannot hold a value of gno.land/r/tests/vm/fvx.E outside its realm (use a pointer)
