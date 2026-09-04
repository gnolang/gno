package store

var admin address

// Fixed: derive the caller from cur.Previous(), which names the realm that
// crossed into this frame. The chain/runtime/unsafe import is gone.
func Set(cur realm, next address) {
	if cur.Previous().Address() != admin {
		panic("admin only")
	}
	admin = next
}
