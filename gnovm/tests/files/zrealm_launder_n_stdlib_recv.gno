// PKGPATH: gno.land/r/laundern
package laundern

// Attack N: the stdlib-receiver analogue of attack M (the /p/-receiver case).
//
// base64.StdEncoding is a stdlib-init-stamped receiver. Under the borrow
// rules, a stdlib-stamped receiver does NOT borrow (borrow rule #2 is skipped
// for stdlib), so base64.Encode runs with the CALLER's realm (/r/laundern),
// not stdlib's. That lets a stdlib method write the caller's OWN out-param
// buffer — but it must NOT let the attacker write a *foreign* realm's buffer.
//
// Here /r/laundern passes /r/laundervictim's own buffer as the Encode dst.
// The write is attributed to /r/laundern, which doesn't own the victim's
// /r/laundervictim-stamped backing array → readonly gate rejects it.

import (
	"encoding/base64"

	"gno.land/r/tests/vm/laundervictim"
)

func main(cur realm) {
	buf := laundervictim.GetBuf() // victim's own /r/laundervictim-stamped buffer
	base64.StdEncoding.Encode(buf, []byte("xx"))
	println("after:", laundervictim.ReadBuf())
}

// Error:
// cannot directly modify readonly tainted object (use a method or crossing function): dst<VPBlock(1,1)>[di<VPBlock(1,3)> + (const (0 int))]
