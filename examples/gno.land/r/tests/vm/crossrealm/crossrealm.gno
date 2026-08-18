package crossrealm

import (
	"chain/runtime"
	"chain/runtime/unsafe"

	"gno.land/p/demo/tests/p_crossrealm"
	"gno.land/p/nt/ownable/v0"
	"gno.land/p/nt/ufmt/v0"
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
	fooer.Foo(cross(cur))
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
	fooerGetter().Foo(cross(cur))
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
// SECURITY XXX: Closure and Closure2 below are exported package-level
// function-typed vars that any foreign realm can set via SetClosure /
// SetClosure2 and trigger via ExecuteClosure / ExecuteClosureCross.
// Closure2's signature is `func(realm)`, which means whoever sets it
// receives a realm value at invocation time — capability-bearing data
// stored in an exported package var, then handed back to caller-supplied
// code. THIS IS DELIBERATE VM-PARITY-TEST INFRASTRUCTURE — these vars
// exist to exercise the VM's handling of stored closures and cross-
// realm callbacks. DO NOT COPY THIS PATTERN IN PRODUCTION CODE. Real
// /r/ realms must never expose `func(realm)` (or any function-typed
// value capable of receiving cur) as an exported package var.
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
	Closure2(cross(cur))
}

// Closure3 mirrors Closure but for non-crossing-with-rlm closures —
// `func(_ int, rlm realm)`. Whoever sets Closure3 receives cur as a
// plain value (no realm boundary at invocation time), then can cross
// internally via cross(rlm).
var Closure3 func(_ int, rlm realm)

func SetClosure3(cur realm, f func(_ int, rlm realm)) {
	Closure3 = f
}

func ExecuteClosure3(cur realm) {
	Closure3(0, cur)
}

// Closure -> FooUpdate
func PrintRealms(cur realm) {
	ufmt.Printf("current realm: %s\n", unsafe.CurrentRealm())
	ufmt.Printf("previous realm: %s\n", unsafe.PreviousRealm())
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
	println("crossrealm  EntryPoint: " + unsafe.PreviousRealm().PkgPath())
	println("crossrealm  EntryPoint: " + unsafe.PreviousRealm().Address())
	println()
	return PrevRealmNoCrossing()
}

// EntryPointWithCrossing is a non-crossing helper that forwards into
// the (cur realm)-crossing PrevRealmCrossing. Callers pass their own
// live cur; `cross(rlm)` does the actual cross.
func EntryPointWithCrossing(_ int, rlm realm) (withCros *ownable.Ownable) {
	return PrevRealmCrossing(cross(rlm))
}

func PrevRealmNoCrossing() *ownable.Ownable {
	println("crossrealm PreviousRealm no crossing: " + unsafe.PreviousRealm().PkgPath())
	println("crossrealm PreviousRealm no crossing: " + unsafe.PreviousRealm().Address())
	return ownable.NewWithAddress(unsafe.CurrentRealm().Address())
}

func PrevRealmCrossing(cur realm) *ownable.Ownable {
	println("crossrealm PreviousRealm with crossing: " + unsafe.PreviousRealm().PkgPath())
	println("crossrealm PreviousRealm with crossing: " + unsafe.PreviousRealm().Address())
	return ownable.NewWithAddress(cur.Address())
}

func CurRealmNoCrossing() runtime.Realm {
	return unsafe.CurrentRealm()
}
func CurRealmCrossing(cur realm) runtime.Realm {
	return unsafe.CurrentRealm()
}

// call the package that returns current realm
func PkgCurRealmNoCrossing() runtime.Realm {
	return p_crossrealm.CurrentRealm()
}

// call the package that returns current realm
func PkgCurRealmCrossing(cur realm) runtime.Realm {
	return unsafe.CurrentRealm()
}
