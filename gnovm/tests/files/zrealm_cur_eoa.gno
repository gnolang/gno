// PKGPATH: gno.land/r/cureoa
package cureoa

import (
	"chain/runtime/unsafe"
)

// Verifies that cur.Previous() at the chain root mirrors what
// unsafe.PreviousRealm() reports for the same context: the per-tx
// origin realm carries OriginCaller as Address.
func main(cur realm) {
	prev := cur.Previous()
	println("cur.Previous() non-nil:", prev != nil)
	println("addr match:", string(prev.Address()) == string(unsafe.PreviousRealm().Address()))
	println("pkgPath match (both empty):", prev.PkgPath() == "" && unsafe.PreviousRealm().PkgPath() == "")
	// Address should be a non-empty EOA-shaped bech32, never empty.
	println("addr non-empty:", string(prev.Address()) != "")
}

// Output:
// cur.Previous() non-nil: true
// addr match: true
// pkgPath match (both empty): true
// addr non-empty: true
