// PKGPATH: gno.land/r/laundero
package laundero

// Attack O: TYPE-PUNNING LAUNDER through a multi-value declaration.
//
// The relabelling variant of attack G. /r/laundervictim hands out its stored
// []byte, and /p/launderattack declares Bytes, a []byte with a Poke mutator.
// `launderattack.Bytes(buf)` is refused by Case 1 in convertValue, since buf
// is stored in another realm. `var b, _ launderattack.Bytes = two()` must be
// refused the same way: otherwise b keeps aliasing the victim's buffer under a
// /p/ type, and Poke runs under borrow rule #2 with the victim's authority.
// two() is the attacker's own; any getter that returns a stored unnamed slice,
// map or pointer can feed it.

import (
	"gno.land/p/demo/tests/launderattack/v0"
	"gno.land/r/tests/vm/laundervictim"
)

func two() ([]byte, []byte) { return laundervictim.GetBuf(), nil }

func main(cur realm) {
	var b, _ launderattack.Bytes = two()
	b.Poke("pwnd")
	println("after:", laundervictim.ReadBuf())
}

// Error:
// illegal conversion of readonly or externally stored value (convert it inside a crossing function of the realm that stores it)
