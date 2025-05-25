package crossrealm

import (
	"std"

	"gno.land/p/demo/ownable"
	"gno.land/p/demo/tests/p_crossrealm"
	"gno.land/p/demo/ufmt"
)

type LocalStruct struct {
	A int
}

func (ls *LocalStruct) String() string {
	return ufmt.Sprintf("LocalStruct{%d}", ls.A)
}

// local is saved locally in this realm
var local *LocalStruct

func init() {
	local = &LocalStruct{A: 123}
}

// Make1 returns a local object wrapped by a p struct
func Make1() *p_crossrealm.Container {
	return &p_crossrealm.Container{
		A: 1,
		B: local,
	}
}

type Fooer interface{ Foo() }

var fooer Fooer

func SetFooer(f Fooer) Fooer {
	crossing()

	fooer = f
	return fooer
}

func GetFooer() Fooer { return fooer }

func CallFooerFoo() { fooer.Foo() }

func CallFooerFooSR() {
	crossing()

	fooer.Foo()
}

func CallFooerFooSR2() {
	crossing()

	cross(fooer.Foo)()
}

type FooerGetter func() Fooer

var fooerGetter FooerGetter

func SetFooerGetter(fg FooerGetter) FooerGetter {
	crossing()

	fooerGetter = fg
	return fg
}

func GetFooerGetter() FooerGetter {
	return fooerGetter
}

func CallFooerGetterFoo() { fooerGetter().Foo() }

func CallFooerGetterFooSR2() {
	crossing()

	cross(fooerGetter().Foo)()
}

// This is a top function that does switch realms.
func ExecSR(cb func() string) string {
	crossing()

	return cb()
}

// This is a top function that doesn't switch realms.
func Exec(cb func() string) string {
	return cb()
}

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

func EntryPoint() (noCros *ownable.Ownable) {
	println("crossrealm  EntryPoint: " + std.PreviousRealm().PkgPath())
	println("crossrealm  EntryPoint: " + std.PreviousRealm().Address())
	println()
	return PrevRealmNoCrossing()
}

func EntryPointWithCrossing() (withCros *ownable.Ownable) {
	return cross(PrevRealmCrossing)()
}

func PrevRealmNoCrossing() *ownable.Ownable {
	println("crossrealm PreviousRealm no crossing: " + std.PreviousRealm().PkgPath())
	println("crossrealm PreviousRealm no crossing: " + std.PreviousRealm().Address())
	return ownable.New()
}

func PrevRealmCrossing() *ownable.Ownable {
	crossing()
	println("crossrealm PreviousRealm with crossing: " + std.PreviousRealm().PkgPath())
	println("crossrealm PreviousRealm with crossing: " + std.PreviousRealm().Address())
	return ownable.New()
}

func CurRealmNoCrossing() std.Realm {
	return std.CurrentRealm()
}
func CurRealmWithCrossing() std.Realm {
	crossing()
	return std.CurrentRealm()
}

// call the package that returns current realm
func PkgCurRealmNoCrossing() std.Realm {
	return p_crossrealm.CurrentRealm()
}

// call the package that returns current realm
func PkgCurRealmWithCrossing() std.Realm {
	crossing()
	return std.CurrentRealm()
}
