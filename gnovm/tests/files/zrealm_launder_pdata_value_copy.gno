// PKGPATH: gno.land/r/launderpdata_v
package launderpdata_v

// /p/-source coverage: direct value-read of a /p/-init-allocated
// package global into a /r/-realm local.
//
// Post-#5747: the value-read deep-copies the /p/-stamped global into
// a fresh, realm-local heap value. The local mutation succeeds and
// has no effect on the /p/-global (the package is re-initialized at
// the start of each transaction anyway). The /p/-immutability gate
// fires on writes that go *through* a /p/-stamped pointer, not on
// writes to local copies.

import (
	"gno.land/p/demo/tests/launderpkg"
)

func main(cur realm) {
	stolen := launderpkg.PInitData
	stolen.Field = "pwnd-via-pdata-value-launder"
	println(stolen.Field)
}

// Output:
// pwnd-via-pdata-value-launder
