package store

import "chain/runtime/unsafe"

var admin address

// Flagged: Set is a crossing function, so it receives cur realm, but derives the
// caller from unsafe.PreviousRealm(). Here that returns the same realm as
// cur.Previous(); what it ignores is the cur token the runtime minted for exactly
// this purpose. Reached from a non-crossing entry point the same call names
// whatever realm crossed into the nearest enclosing crossing function, not the
// immediate caller. The cur parameter is accepted but never used.
func Set(cur realm, next address) {
	if unsafe.PreviousRealm().Address() != admin {
		panic("admin only")
	}
	admin = next
}
