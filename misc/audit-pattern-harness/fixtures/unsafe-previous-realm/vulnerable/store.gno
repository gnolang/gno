package store

import "chain/runtime/unsafe"

var admin address

// Vulnerable: Set is a crossing function, so it receives cur realm, but derives
// the caller from unsafe.PreviousRealm(). That walks the stack to the outermost
// crossing realm instead of naming the immediate caller, and ignores the cur
// token the runtime minted for exactly this purpose. The cur parameter is
// accepted but never used.
func Set(cur realm, next address) {
	if unsafe.PreviousRealm().Address() != admin {
		panic("admin only")
	}
	admin = next
}
