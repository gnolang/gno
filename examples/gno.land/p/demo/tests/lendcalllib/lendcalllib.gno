// Package lendcalllib holds library code the lend filetests call.
package lendcalllib

var n int
var bump func()

// The library's own init writes its state through a function value.
func init() {
	bump = func() { n++ }
	bump()
	bump()
}

func N() int { return n }

func Call(f func()) { f() }

// Ret calls f where it is written and returns what it got.
func Ret(f func() int) int { return f() }

// Spot builds a literal and calls it on the spot, inside the library.
func Spot(f func()) { func() { f() }() }

// Hold calls f and keeps it.
func Hold(f func(), box *[]func()) { f(); *box = append(*box, f) }

// DeferCall defers f: it runs before DeferCall returns.
func DeferCall(f func()) { defer f() }

// KeepX keeps a crossing function value in one slot (no variadic append).
func KeepX(g func(realm), box *func(realm)) { *box = g }

// Relay passes g on to KeepX: g is lent to Relay, then kept.
func Relay(g func(realm), box *func(realm)) { KeepX(g, box) }
