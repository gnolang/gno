// PKGPATH: gno.land/r/hivrealrejected
package hivrealrejected

// Negative test for the persisted-closure-capture case.
//
// persistedcap's init() calls p_closurecap.MakeCounter, which returns
// a closure capturing a local `c`. The closure is stored at package
// level in persistedcap and persisted as part of /r/.../persistedcap's
// state — the captured HIV has PkgID = persistedcap, NewTime > 0
// (real).
//
// From this realm (hivrealrejected), we grab the closure via
// GetCounter() and invoke it directly. The closure's FuncLit lives in
// /p/ p_closurecap, so PushFrameCall does NOT switch m.Realm to
// persistedcap — m.Realm stays hivrealrejected. The write to `c`
// inside the closure body therefore sees m.Realm = hivrealrejected,
// HIV.PkgID = persistedcap, and the HIV is real. The unreal-HIV
// exception in IsReadonlyBy doesn't apply, so the standard PkgID
// gate fires and the write is rejected.
//
// Cross-realm side effects must be syntactically visible via `cross`:
// to share counter state across realms, persistedcap should expose a
// crossing function like `func Increment(cur realm) int` instead of
// handing back the raw closure.

import "gno.land/r/tests/vm/persistedcap"

func main(cur realm) {
	fn := persistedcap.GetCounter()
	fn()
}

// Error:
// cannot directly modify readonly tainted object (use a method or crossing function): c<~VPBlock(1,1)>
