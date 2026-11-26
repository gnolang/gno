// Package t73leak hands its `cur` out of a call in every way the
// preprocess rule does not see, so tests can show what the holder can do.
package t73leak

import "chain/banker"

var (
	Count int
	Held  realm
)

// Bump writes as this realm; a stale `cur` must not reach it via cross.
func Bump(cur realm) { Count++ }

func LeakDecl(cur realm) realm { return cur }

func LeakAny(cur realm) any { return func() any { return cur }() }

type Box struct{ R realm }

func LeakField(cur realm) *Box { return &Box{R: cur} }

func LeakPanic(cur realm) { panic(cur) }

func LeakPkgVar(cur realm) { Held = cur }

func LeakCaptured(cur realm, set func(int, realm)) { set(0, cur) }

func LeakDefer(cur realm, sink func(int, realm)) { defer sink(0, cur) }

// LeakClosure returns a closure that uses `cur` after this call ended.
func LeakClosure(cur realm) func(string) {
	return func(what string) {
		switch what {
		case "PkgPath":
			println("closure PkgPath:", cur.PkgPath())
		case "banker":
			banker.NewBanker(banker.BankerTypeRealmSend, cur)
		case "cross":
			Bump(cross(cur))
		case "Sub":
			banker.NewBanker(banker.BankerTypeRealmSend, cur.Sub("x"))
		}
	}
}
