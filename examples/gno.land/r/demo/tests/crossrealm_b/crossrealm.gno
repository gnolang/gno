package crossrealm_b

import (
	"std"

	"gno.land/r/demo/tests/crossrealm"
)

type fooer struct {
	s string
}

func (f *fooer) SetS(newVal string) {
	f.s = newVal
}

func (f *fooer) Foo() {
	crossing()

	println("hello " + f.s + " cur=" + std.CurrentRealm().PkgPath() + " prev=" + std.PreviousRealm().PkgPath())
}

var (
	Fooer              = &fooer{s: "A"}
	FooerGetter        = func() crossrealm.Fooer { return Fooer }
	FooerGetterBuilder = func() crossrealm.FooerGetter { return func() crossrealm.Fooer { return Fooer } }
)

var Closure func()

func SetClosure(f func()) {
	crossing()

	Closure = f
}

var Object any

func SetObject(x any) {
	crossing()

	Object = x
}

func GetObject() any {
	return Object
}
