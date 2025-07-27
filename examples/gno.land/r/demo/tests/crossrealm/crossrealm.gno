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

type Fooer interface {
	Foo(realm)
	Bar()
}

var fooer Fooer

func SetFooer(cur realm, f Fooer) Fooer {
	fooer = f
	return fooer
}

func GetFooer() Fooer {
	return fooer
}

func CallFooerFooCur(cur realm) {
	fooer.Foo(cur)
}

func CallFooerFooCross(cur realm) {
	fooer.Foo(cross)
}

func CallFooerBar() {
	fooer.Bar()
}

func CallFooerBarCrossing(cur realm) {
	fooer.Bar()
}

type FooerGetter func() Fooer

var fooerGetter FooerGetter

func SetFooerGetter(cur realm, fg FooerGetter) FooerGetter {
	fooerGetter = fg
	return fg
}

func GetFooerGetter() FooerGetter {
	return fooerGetter
}

func CallFooerGetterBar() {
	fooerGetter().Bar()
}

func CallFooerGetterBarCrossing(cur realm) {
	fooerGetter().Bar()
}

func CallFooerGetterFooCur(cur realm) {
	fooerGetter().Foo(cur)
}

func CallFooerGetterFooCross(cur realm) {
	fooerGetter().Foo(cross)
}

// This is a top function that does switch realms.
func ExecCrossing(cur realm, cb func() string) string {
	return cb()
}

// This is a top function that doesn't switch realms.
func Exec(cb func() string) string {
	return cb()
}

// ------------------------------------
var AllowedList = []int{1, 2}
var Closure func()

func SetClosure(cur realm, f func()) {
	Closure = f
}

func ExecuteClosure(cur realm) {
	Closure()
}

var Closure2 func(realm)

func SetClosure2(cur realm, f func(realm)) {
	Closure2 = f
}
func ExecuteClosureCross(cur realm) {
	Closure2(cross)
}

// Closure -> FooUpdate
func PrintRealms(cur realm) {
	ufmt.Printf("current realm: %s\n", std.CurrentRealm())
	ufmt.Printf("previous realm: %s\n", std.PreviousRealm())
}

// -------------------------------------------------
var Object any

func SetObject(cur realm, x any) {
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
	return PrevRealmCrossing(cross)
}

func PrevRealmNoCrossing() *ownable.Ownable {
	println("crossrealm PreviousRealm no crossing: " + std.PreviousRealm().PkgPath())
	println("crossrealm PreviousRealm no crossing: " + std.PreviousRealm().Address())
	return ownable.New()
}

func PrevRealmCrossing(cur realm) *ownable.Ownable {
	println("crossrealm PreviousRealm with crossing: " + std.PreviousRealm().PkgPath())
	println("crossrealm PreviousRealm with crossing: " + std.PreviousRealm().Address())
	return ownable.New()
}

func CurRealmNoCrossing() std.Realm {
	return std.CurrentRealm()
}
func CurRealmCrossing(cur realm) std.Realm {
	return std.CurrentRealm()
}

// call the package that returns current realm
func PkgCurRealmNoCrossing() std.Realm {
	return p_crossrealm.CurrentRealm()
}

// call the package that returns current realm
func PkgCurRealmCrossing(cur realm) std.Realm {
	return std.CurrentRealm()
}
