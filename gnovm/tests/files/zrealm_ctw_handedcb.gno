// PKGPATH: gno.land/r/ctw_handedcb
// A realm's closure handed to another realm still runs read-only there: the
// nearest realm frame is the holder's, not the owner's.
package ctw_handedcb

import (
	"gno.land/p/demo/tests/libtable/v0"
	"gno.land/r/tests/vm/ctw"
)

func main(cur realm) {
	libtable.Run(ctw.Hook())
}

// Error:
// gno.land/r/tests/vm/ctw's code was entered without crossing and is read-only (call its crossing function): (const (ref(gno.land/r/tests/vm/ctw) package{})).Count
