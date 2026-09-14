package store

var admin address

// Fixed: derive the caller from cur.Previous() under a cur.IsCurrent() guard,
// which verifies the live crossing frame. The chain/runtime/unsafe import is
// gone.
func Set(cur realm, next address) {
	if !cur.IsCurrent() {
		panic("invalid realm")
	}
	if cur.Previous().Address() != admin {
		panic("admin only")
	}
	admin = next
}
