package cbvictim

import (
	"gno.land/p/demo/tests/launderpkg/v0"
	"gno.land/p/demo/tests/libtable/v0"
	"gno.land/p/demo/tests/pfuncbind"
)

var Owner = "victim"

// SetOwner is exported and not crossing: callers outside must not write.
func SetOwner(s string) { Owner = s }

// Run is crossing and runs a function value the caller handed in.
func Run(cur realm, f func()) { f() }

// RunS is crossing and runs a function value the caller handed in.
func RunS(cur realm, f func(string)) { f("picked") }

// Getter returns a closure; made in read-only code it carries no authority.
func Getter() func(string) { return func(s string) { Owner = s } }

// Keep is a closure made in a gate: it writes for this realm.
var Keep func(string)

func Init(cur realm) { Keep = func(s string) { Owner = s } }

type T struct{ Name string }

var Box = &T{}

func (t *T) SetName(s string) { t.Name = s }

var hook func(string)

func SetHook(cur realm, f func(string)) { hook = f }
func RunHook(cur realm)                 { hook("stored") }

// Writers builds a closure writing Owner and one writing what it is given;
// entered without crossing, both carry no maker.
func Writers() (func(string), func([]string)) {
	return func(s string) { Owner = s }, func(b []string) { b[0] = "pwnd" }
}

// GetMV returns a method value bound while read-only: it has no binder.
func GetMV() func(string) { return Box.SetName }

type S struct{ F func(string) }

// GetSt returns a struct holding a method value bound while read-only.
func GetSt() *S { return &S{F: Box.SetName} }

// RunSt is crossing and runs the function a caller's struct holds.
func RunSt(cur realm, s *S) { s.F("struct") }

// GetDecl and FV hand out this realm's own declared function.
func GetDecl() func(string) { return SetOwner }

var FV = SetOwner

type V struct{ P *T }

func (v V) SetName(s string) { v.P.Name = s }

var Val = V{P: Box}

type Setter interface{ SetName(string) }

// Bound while read-only: no binder.
func GetValMV() func(string)   { return Val.SetName }
func GetIfaceMV() func(string) { var s Setter = Box; return s.SetName }

// Bound inside a crossing function: writes for this realm.
func GateValMV(cur realm) func(string)   { return Val.SetName }
func GateIfaceMV(cur realm) func(string) { var s Setter = Box; return s.SetName }

// Get2 hands out SetOwner from inside a closure.
func Get2() func(string) { return func() func(string) { return SetOwner }() }

// Get3 hands out SetOwner through a variable a closure assigns.
func Get3() func(string) {
	var f func(string)
	func() { f = SetOwner }()
	return f
}

// GateDecl hands out SetOwner from inside a gate: it writes for this realm.
func GateDecl(cur realm) func(string) { return SetOwner }

// GetMExpr and GateMExpr hand out a method expression, read-only and in a gate.
func GetMExpr() func(*T, string)           { return (*T).SetName }
func GateMExpr(cur realm) func(*T, string) { return (*T).SetName }

// RunM is crossing and runs a method expression the caller handed in on Box.
func RunM(cur realm, f func(*T, string)) { f(Box, "picked") }

// Lib is an object of a library type this realm owns.
var Lib = &launderpkg.Immutable{Field: "orig"}

// LibGate calls a library writer through a value made in this gate.
func LibGate(cur realm) {
	f := pfuncbind.Clobber
	f(Lib)
}

// GetLib hands out a library writer, made read-only.
func GetLib() func(*launderpkg.Immutable) { return pfuncbind.Clobber }

// LibFV is a library writer taken as a value at init.
var LibFV = pfuncbind.Clobber

// RunLib is crossing and runs a library-typed callback on Lib.
func RunLib(cur realm, f func(*launderpkg.Immutable)) { f(Lib) }

// LibGateless calls a library writer through a value, without a gate.
func LibGateless() {
	f := pfuncbind.Clobber
	f(Lib)
}

// Values made at init: maker cbvictim. A gate below runs each one.
var (
	Table  = []func(string){SetOwner}
	Map    = map[string]func(string){"a": SetOwner}
	Arr    = [1]func(string){SetOwner}
	Obj    = &S{F: SetOwner}
	Nested = Pair{F: SetOwner}
	H      = Hook(SetOwner)
	MV     = Box.SetName
	Kick   = func() { Owner = "kick" }
)

// Pair is an all-exported struct whose copies may leave the realm.
type Pair = struct{ F func(string) }

// Hook is a function type with a crossing method that calls its receiver.
type Hook func(string)

func (h Hook) Run(cur realm) { h("recv") }

func RunAll(cur realm, fs []func(string)) {
	for _, f := range fs {
		f("table")
	}
}
func RunMap(cur realm, m map[string]func(string)) { m["a"]("map") }
func RunP(cur realm, p *func(string))             { (*p)("ptr") }
func RunAny(cur realm, x any)                     { x.(*S).F("any") }
func RunIface(cur realm, x any)                   { x.(func(string))("iface") }
func RunArr(cur realm, a [1]func(string))         { a[0]("arr") }
func RunNested(cur realm, p Pair)                 { p.F("nested") }
func RunGet(cur realm, get func() func(string))   { get()("provider") }

// RunOwnTable, RunOwnHelper and RunLibIterate run this realm's own values.
func RunOwnTable(cur realm) {
	for _, f := range Table {
		f("own-table")
	}
}
func RunOwnHelper(cur realm)         { apply(FV, "own-helper") }
func apply(f func(string), s string) { f(s) }
func RunLibIterate(cur realm)        { libtable.Run(Kick) }

// SetLibHook keeps a library-coded value a caller hands in.
var libHook func(*launderpkg.Immutable)

func SetLibHook(cur realm, f func(*launderpkg.Immutable)) { libHook = f }

// SetX is a crossing closure made at init; another realm may register it.
var SetX = func(cur realm, s string) { Owner = s }

// GetFV hands out FV from read-only code.
func GetFV() func(string) { return FV }

// Values nested by value: a struct in a struct, behind a pointer, arrays.
type Outer = struct{ Inner Pair }

var (
	O  = Outer{Inner: Pair{F: SetOwner}}
	P  = &Outer{Inner: Pair{F: SetOwner}}
	A  = [1]Pair{{F: SetOwner}}
	AA = [1][1]func(string){{SetOwner}}
	AS = [1][1]string{{"old"}}
)

func RunOuter(cur realm, o Outer)               { o.Inner.F("outer") }
func RunArrPair(cur realm, a [1]Pair)           { a[0].F("arr") }
func RunArrArr(cur realm, a [1][1]func(string)) { a[0][0]("arrarr") }

// CheckAA runs AA's own element; SetAS changes AS through this realm's gate.
func CheckAA(cur realm) { AA[0][0]("set-intact") }
func SetAS(cur realm)   { AS[0][0] = "new" }

// RunPanic recovers a value the caller's callback panics with and runs it.
func RunPanic(cur realm, cb func()) {
	defer func() {
		if f, ok := recover().(func(string)); ok {
			f("panic")
		}
	}()
	cb()
}

// RegVia registers FV in a registry realm's gates, takes it back and runs it.
func RegVia(cur realm, put func(cur realm, f func(string)), get func(cur realm) func(string)) {
	put(cross(cur), FV)
	f := get(cross(cur))
	f("registry")
}
