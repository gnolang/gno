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
