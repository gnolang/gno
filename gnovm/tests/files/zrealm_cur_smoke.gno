// PKGPATH: gno.land/r/curtest
package curtest

import (
	"chain/runtime/unsafe"
)

func Stash(_ int, rlm realm) realm {
	rlm_ := rlm
	println("cur == cur_:", rlm == rlm_)
	println("cur.Address:", rlm.Address())
	println("cur.PkgPath:", rlm.PkgPath())
	println("runtime.Current matches:", string(rlm.Address()) == string(unsafe.CurrentRealm().Address()))
	// At the chain root the prev is the EOA origin: non-nil pointer with
	// empty pkgPath (mirrors unsafe.PreviousRealm()'s convention). Calling
	// .Previous() past the origin panics (mirrors unsafe.PreviousRealm()'s
	// walk-end behavior), so user code checks PkgPath()=="" to stop.
	prev := rlm.Previous()
	println("Previous() at top non-nil:", prev != nil)
	println("Previous().PkgPath empty:", prev.PkgPath() == "")
	func() {
		defer func() {
			r := recover()
			println("Previous().Previous() panics past origin:", r != nil)
		}()
		_ = prev.Previous()
	}()
	return rlm
}

func main(cur realm) {
	r := Stash(0, cur)
	println("returned non-nil:", r != nil)
}

// Output:
// cur == cur_: true
// cur.Address: g1z224xsf85kygxt6nvyxl2xy7a8yuj0tl6taj59
// cur.PkgPath: gno.land/r/curtest
// runtime.Current matches: true
// Previous() at top non-nil: true
// Previous().PkgPath empty: true
// Previous().Previous() panics past origin: true
// returned non-nil: true
