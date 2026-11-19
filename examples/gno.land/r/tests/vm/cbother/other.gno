// Package cbother is a third realm next to cbvictim: it relays and keeps
// cbvictim's function values.
package cbother

import "gno.land/r/tests/vm/cbvictim"

// Relay hands a value it read from cbvictim back to cbvictim's gate.
func Relay(cur realm) { cbvictim.RunS(cross(cur), cbvictim.Table[0]) }

// RelayTable hands cbvictim's own table back by reference.
func RelayTable(cur realm) { cbvictim.RunAll(cross(cur), cbvictim.Table) }

// Kept is a value this realm keeps; RunKept calls it.
var Kept func(string)

func Keep(cur realm, f func(string)) { Kept = f }
func KeepOwn(cur realm)              { Kept = cbvictim.FV }
func RunKept(cur realm)              { Kept("kept") }

// Ref keeps a reference to cbvictim's table; HandRef passes it back.
var Ref []func(string)

func TakeRef(cur realm) { Ref = cbvictim.Table }
func HandRef(cur realm) { cbvictim.RunAll(cross(cur), Ref) }

// handle holds crossing functions another realm registers (grc20reg shape).
type handle struct{ set func(cur realm, s string) }

var handles []*handle

func Register(cur realm, set func(cur realm, s string)) {
	handles = append(handles, &handle{set: set})
}

func CallAll(cur realm) {
	for _, h := range handles {
		h.set(cross(cur), "crossed")
	}
}

// Each route hands cbvictim's nested value to its gate; warm reads the nested
// function first, loading it in the same transaction.
func Outer(cur realm, warm bool) {
	if warm && cbvictim.O.Inner.F == nil {
		panic("nil")
	}
	cbvictim.RunOuter(cross(cur), cbvictim.O)
}

func Ptr(cur realm, warm bool) {
	if warm && cbvictim.P.Inner.F == nil {
		panic("nil")
	}
	cbvictim.RunOuter(cross(cur), *cbvictim.P)
}

func Arr(cur realm, warm bool) {
	if warm && cbvictim.A[0].F == nil {
		panic("nil")
	}
	cbvictim.RunArrPair(cross(cur), cbvictim.A)
}

func ArrArr(cur realm, warm bool) {
	if warm && cbvictim.AA[0][0] == nil {
		panic("nil")
	}
	cbvictim.RunArrArr(cross(cur), cbvictim.AA)
}

// keepAA and keepAS are this realm's copies of cbvictim's arrays.
var (
	keepAA [1][1]func(string)
	keepAS [1][1]string
)

func KeepAA(cur realm) { keepAA = cbvictim.AA }

func PassAA(cur realm, warm bool) {
	if warm && keepAA[0][0] == nil {
		panic("nil")
	}
	cbvictim.RunArrArr(cross(cur), keepAA)
}

func MutAA(cur realm)         { keepAA[0][0] = func(s string) {} }
func KeepAS(cur realm)        { keepAS = cbvictim.AS }
func MutAS(cur realm)         { keepAS[0][0] = "mine" }
func ReadAS(cur realm) string { return keepAS[0][0] }
