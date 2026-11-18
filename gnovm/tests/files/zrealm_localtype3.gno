// PKGPATH: gno.land/r/test
package test

type T struct{ x int }

func (t T) Get() int { return t.x }

type Getter interface{ Get() int }

// Addpkg-time escapes: X (file-level var init) and Y (init func) persist in
// the addpkg tx itself, pinning the saveFuncLocalTypes-before-finalization
// order. main only reads (bare Realm golden); fails on master reloading X.
func mk() any {
	type S struct{ T }
	return S{T{9}}
}

var (
	X any = mk()
	Y any
)

func init() {
	type U struct{ T }
	Y = U{T{4}}
}

func main(cur realm) {
	println(X.(Getter).Get(), Y.(Getter).Get())
}

// Output:
// 9 4

// Realm:
// finalizerealm["gno.land/r/test"]
