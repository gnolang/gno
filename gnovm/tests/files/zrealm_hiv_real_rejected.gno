// PKGPATH: gno.land/r/hivrealrejected
package hivrealrejected

// Persisted-closure-capture case under borrow rule #3 (closure capture-realm
// borrow).
//
// persistedcap's init() calls p_closurecap.MakeCounter, which returns
// a closure capturing a local `c`. The closure is stored at package
// level in persistedcap and persisted as part of /r/.../persistedcap's
// state — the captured HIV has PkgID = persistedcap, NewTime > 0
// (real). The closure's FuncValue itself is stamped PkgID =
// persistedcap at construction time.
//
// From this realm (hivrealrejected), we grab the closure via
// GetCounter() and invoke it directly. The closure's FuncLit lives in
// /p/ p_closurecap, so borrow rule #1 (declaring-realm borrow) doesn't fire.
// No receiver, so borrow rule #2 doesn't fire either. borrow rule #3 fires: the
// closure's stamped PkgID (persistedcap) differs from m.Realm
// (hivrealrejected), so m.Realm is borrowed to persistedcap for the
// closure body. The write to `c` then runs in-realm under
// persistedcap's authority — the persisted counter advances.

import "gno.land/r/tests/vm/persistedcap"

func main(cur realm) {
	fn := persistedcap.GetCounter()
	println(fn())
	println(fn())
}

// Output:
// 1
// 2
