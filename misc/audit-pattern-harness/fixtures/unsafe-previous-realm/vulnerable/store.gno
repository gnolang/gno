package store

import "chain/runtime/unsafe"

var admin address

// Vulnerable: Set is a crossing function — it receives cur realm — but derives
// the caller from unsafe.PreviousRealm(). That skips the cur.IsCurrent() frame
// verification that makes cur.Previous() safe and ignores the cur token the
// runtime minted for exactly this purpose. The cur parameter is accepted but
// never used.
func Set(cur realm, next address) {
	if unsafe.PreviousRealm().Address() != admin {
		panic("admin only")
	}
	admin = next
}
