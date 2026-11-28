// PKGPATH: gno.land/r/laundero
package laundero

// A multi-value declaration that drops another realm's stored buffer into `_`
// converts nothing for it, as `_, m = pair()` does: the dropped value keeps no
// alias, so there is nothing to refuse. The kept result still converts.

import "gno.land/r/tests/vm/laundervictim"

type Mine []byte

func pair() ([]byte, []byte) { return laundervictim.GetBuf(), []byte("own") }

func main(cur realm) {
	var m Mine
	_, m = pair()
	println("assign", string(m))
	var _, n Mine = pair()
	println("decl", string(n))
}

// Output:
// assign own
// decl own
