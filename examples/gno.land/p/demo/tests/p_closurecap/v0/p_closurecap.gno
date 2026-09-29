// Package p_closurecap exercises the "var stopped bool in
// memberStorage.IterateByOffset" pattern from boards2/commondao:
// a method declares a local var that an inline closure captures and
// writes to. Without the unreal-HIV exception in the readonly check,
// this failed across borrow-realm transitions because the HIV's PkgID
// stamp records the alloc-site realm, but the closure body may run
// under a different borrowed realm — the check would fire on a write
// the closure itself made to its own captured slot.
package p_closurecap

// Inner is the receiver of loop(). When Inner lives in a different
// realm than the Outer-method's caller realm, PushFrameCall's borrow
// rule 2 shifts m.Realm to Inner's realm for the entire body of loop()
// (and the synchronously-invoked closure). That shift is what makes
// HIV.PkgID (stamped at var stopped's alloc) differ from m.Realm at
// the closure's write site.
type Inner struct {
	N int
}

// Outer is the var-stopped pattern. The inline closure captures
// `stopped` and writes to it. Returns true if `fn` ever returned true.
//
// The key shape: Outer is a top-level /p/ func (no receiver, so no
// borrow on entry — HIV stamps with caller's realm), but it delegates
// to `inn.loop(...)`, where `inn` is supplied by the caller and may
// live in a different realm. inn.loop's borrow flips m.Realm to inn's
// realm; the inline closure then writes to `stopped` under that
// borrowed realm.
func Outer(inn *Inner, count int, fn func(i int) bool) bool {
	var stopped bool
	inn.loop(count, func(i int) bool {
		stopped = fn(i)
		return stopped
	})
	return stopped
}

// loop is a /p/ method on *Inner. As a receiver-method on a /p/ type,
// PushFrameCall's borrow rule 2 fires here: m.Realm becomes
// Inner.PkgID's realm for the duration of loop().
func (in *Inner) loop(count int, cb func(i int) bool) {
	for i := 0; i < count; i++ {
		if cb(i) {
			return
		}
	}
}

// MakeCounter returns a closure that captures a local int. The
// returned closure can be stored in /r/ state — verifying that a
// persisted closure-capture HIV (whose FuncLit lives in /p/) is still
// writable when invoked from a foreign realm context.
func MakeCounter(start int) func() int {
	c := start
	return func() int {
		c++
		return c
	}
}
