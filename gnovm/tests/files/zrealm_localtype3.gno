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
// main only reads, so its Realm golden is bare (finalizerealm only): the
// interesting persistence happens at addpkg/init, whose store ops the
// filetest framework clears before running main. This test is also the one
// release-build in-process reproducer: without the fix, main's transaction
// reloads X and panics with "unexpected type with id ...".
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
