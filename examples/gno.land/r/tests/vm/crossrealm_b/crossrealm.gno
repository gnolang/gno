package crossrealm_b

import (
	"chain/banker"
	"chain/runtime/unsafe"

	"gno.land/r/tests/vm/crossrealm"
)

type fooer struct {
	s string
}

func (f *fooer) SetS(newVal string) {
	f.s = newVal
}

func (f *fooer) Foo(cur realm) {
	println("hello " + f.s + " cur=" + unsafe.CurrentRealm().PkgPath() + " prev=" + unsafe.PreviousRealm().PkgPath())
}

func (f *fooer) Bar() {
	println("hello " + f.s + " cur=" + unsafe.CurrentRealm().PkgPath() + " prev=" + unsafe.PreviousRealm().PkgPath())
}

var (
	Fooer              = &fooer{s: "A"}
	FooerGetter        = func() crossrealm.Fooer { return Fooer }
	FooerGetterBuilder = func() crossrealm.FooerGetter { return func() crossrealm.Fooer { return Fooer } }
)

var Closure func()

func SetClosure(cur realm, f func()) {
	Closure = f
}

var Object any

func SetObject(cur realm, x any) {
	Object = x
}

func GetObject() any {
	return Object
}

func IncrementObject(cur realm) any {
	ptr := Object.(*int)
	*ptr += 1
	return Object
}

var n int

// NOTE should be non-crossing
func IncrGlobal() {
	n++
}

// TrySubOn attempts to mint a sub-realm token on a passed-in realm
// value. Non-crossing (`_ int` discriminator), so the caller's cur
// remains the topmost crossing cur — but borrow rule #1 runs this
// body with m.Realm = crossrealm_b, so rlm.Sub must reject: a foreign
// realm must not mint sub-identities in the caller's namespace.
func TrySubOn(_ int, rlm realm) {
	rlm.Sub("stolen")
}

// TryBankerOnPrevious attempts to construct a RealmSend banker over the
// CALLER via cur.Previous() (which is NOT IsCurrent). NewBanker must
// reject it — otherwise a callee could set pkgAddr to the caller's
// address and drain it through the pkgAddr==from gate. Regression guard
// for the IsCurrent check in NewBanker.
func TryBankerOnPrevious(cur realm) {
	banker.NewBanker(banker.BankerTypeRealmSend, cur.Previous())
}
