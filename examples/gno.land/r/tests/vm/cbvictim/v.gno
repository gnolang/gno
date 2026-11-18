package cbvictim

import (
	"gno.land/p/demo/tests/launderpkg/v0"
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
