// PKGPATH: gno.land/r/curtest
package curtest

import (
	"chain/runtime"
)

func Stash(cur realm) realm {
	cur_ := cur
	println("cur == cur_:", cur == cur_)
	println("cur.Address:", cur.Address())
	println("cur.PkgPath:", cur.PkgPath())
	println("runtime.Current matches:", string(cur.Address()) == string(runtime.CurrentRealm().Address()))
	// At the chain root the prev is the EOA origin: non-nil pointer with
	// empty pkgPath (mirrors runtime.PreviousRealm()'s convention). Calling
	// .Previous() past the origin panics (mirrors runtime.PreviousRealm()'s
	// walk-end behavior), so user code checks PkgPath()=="" to stop.
	prev := cur.Previous()
	println("Previous() at top non-nil:", prev != nil)
	println("Previous().PkgPath empty:", prev.PkgPath() == "")
	func() {
		defer func() {
			r := recover()
			println("Previous().Previous() panics past origin:", r != nil)
		}()
		_ = prev.Previous()
	}()
	return cur
}

func main() {
	r := Stash(cross)
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
