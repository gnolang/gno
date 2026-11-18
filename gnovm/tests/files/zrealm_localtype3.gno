// PKGPATH: gno.land/r/test
package test

type T struct{ x int }

func (t T) Get() int { return t.x }

type Getter interface{ Get() int }

// Addpkg-time escapes: these local-typed values are persisted by the addpkg
// transaction itself — X by a file-level var initializer (saved by
// saveNewPackageValuesAndTypes, which is why saveFuncLocalTypes must run
// before FinalizeRealmTransaction; under -tags debugAssert the SetObject
// assert fires if that order is broken), Y by init (saved by
// resavePackageValues). All other localtype tests escape via a later call.
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
