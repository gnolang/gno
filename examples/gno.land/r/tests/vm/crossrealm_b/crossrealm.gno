package crossrealm_b

import (
	"std"

	"gno.land/r/tests/vm/crossrealm"
)

type fooer struct {
	s string
}

func (f *fooer) SetS(newVal string) {
	f.s = newVal
}

func (f *fooer) Foo(cur realm) {
	println("hello " + f.s + " cur=" + std.CurrentRealm().PkgPath() + " prev=" + std.PreviousRealm().PkgPath())
}

func (f *fooer) Bar() {
	println("hello " + f.s + " cur=" + std.CurrentRealm().PkgPath() + " prev=" + std.PreviousRealm().PkgPath())
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
