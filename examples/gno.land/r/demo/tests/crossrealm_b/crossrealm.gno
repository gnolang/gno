package crossrealm_b

import (
	"chain/runtime"

	"gno.land/r/demo/tests/crossrealm"
)

type fooer struct {
	s string
}

func (f *fooer) SetS(newVal string) {
	f.s = newVal
}

func (f *fooer) Foo() {
	println("hello " + f.s + " cur=" + runtime.CurrentRealm().PkgPath() + " prev=" + runtime.PreviousRealm().PkgPath())
}

var (
	Fooer              = &fooer{s: "A"}
	FooerGetter        = func() crossrealm.Fooer { return Fooer }
	FooerGetterBuilder = func() crossrealm.FooerGetter { return func() crossrealm.Fooer { return Fooer } }
)
