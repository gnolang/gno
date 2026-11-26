// PKGPATH: gno.land/r/zrealm_fvshare_recover
// A writer that left by a panic — thrown by read-only or by writable code,
// called by a stranger's library or recovered by the caller's realm — never
// writes; Private, a value called from a local slice, neither.
package zrealm_fvshare_recover

import (
	"gno.land/p/demo/tests/fvpass"
	"gno.land/r/tests/vm/fvshare"
)

type R struct{}

func (r *R) Run(f func()) { f() }

// K is a stranger's catcher: it keeps what it recovered behind a getter
// closure and hands it back from Pick.
type K struct {
	put func(g func() any)
	get func() any
}

func (k *K) Catch() {
	v := recover()
	k.put(func() any { return v })
}

func (k *K) Pick() func() {
	f, _ := k.get().(func())
	return f
}

func newK() *K {
	var stash func() any
	return &K{
		put: func(g func() any) { stash = g },
		get: func() any {
			if stash == nil {
				return nil
			}
			return stash()
		},
	}
}

var kept func()

func keepThrow(cur realm) any {
	return revive(func() { fvshare.Throw(cross(cur)) })
}

func try(name string, f func()) {
	r := revive(f)
	println(name, "refused:", r != nil, "count:", fvshare.Count(), "why:", r)
}

// tryCall catches a direct call into the other realm: it never crosses, so
// revive does not see it and the read-only refusal is a plain recover.
func tryCall(name string, f func()) {
	var r any
	func() {
		defer func() { r = recover() }()
		f()
	}()
	println(name, "refused:", r != nil, "count:", fvshare.Count(), "why:", r)
}

func main(cur realm) {
	h := &R{}
	try("recovered in call", func() { fvshare.RecoverRun(cross(cur), h) })
	kept = fvshare.RecoverKeep(cross(cur), h)
	println("kept is set:", kept != nil)
	try("kept after return", func() { fvshare.Call(cross(cur), kept) })
	try("pick after catch", func() { fvshare.ChalPick(cross(cur), newK()) })
	try("pick after lib catch", func() { fvshare.LibPick(cross(cur), fvpass.L{}) })
	for _, shape := range []string{"direct", "struct", "slice"} {
		try("thrown writable, lib calls "+shape, func() { fvshare.ThrowLib(cross(cur), fvpass.L{}, shape) })
		try("thrown writable, lib rethrows "+shape, func() { fvshare.ThrowBack(cross(cur), fvpass.Rethrow{}, shape) })
	}
	try("thrown writable, lib rethrows getter", func() { fvshare.ThrowBack(cross(cur), fvpass.Rethrow{}, "getter") })
	thrown := revive(func() { fvshare.Throw(cross(cur)) })
	tryCall("thrown writable, realm recovers", thrown.(func()))
	keptAfter := keepThrow(cross(cur))
	tryCall("thrown writable, realm keeps", keptAfter.(func()))
	try("private", func() { fvshare.Private(cross(cur)) })
}

// Output:
// recovered in call refused: true count: 0 why: gno.land/r/tests/vm/fvshare's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): v<VPBlock(1,0)>.N
// kept is set: true
// kept after return refused: true count: 0 why: gno.land/r/tests/vm/fvshare's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): v<VPBlock(1,0)>.N
// pick after catch refused: true count: 0 why: gno.land/r/tests/vm/fvshare's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): v<VPBlock(1,0)>.N
// pick after lib catch refused: true count: 0 why: gno.land/r/tests/vm/fvshare's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): v<VPBlock(1,0)>.N
// thrown writable, lib calls direct refused: true count: 0 why: gno.land/r/tests/vm/fvshare's code was entered without crossing and is read-only (call its crossing function): v<VPBlock(1,0)>.N
// thrown writable, lib rethrows direct refused: true count: 0 why: gno.land/r/tests/vm/fvshare's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): v<VPBlock(1,0)>.N
// thrown writable, lib calls struct refused: true count: 0 why: gno.land/r/tests/vm/fvshare's code was entered without crossing and is read-only (call its crossing function): v<VPBlock(1,0)>.N
// thrown writable, lib rethrows struct refused: true count: 0 why: gno.land/r/tests/vm/fvshare's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): v<VPBlock(1,0)>.N
// thrown writable, lib calls slice refused: true count: 0 why: gno.land/r/tests/vm/fvshare's code was entered without crossing and is read-only (call its crossing function): v<VPBlock(1,0)>.N
// thrown writable, lib rethrows slice refused: true count: 0 why: gno.land/r/tests/vm/fvshare's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): v<VPBlock(1,0)>.N
// thrown writable, lib rethrows getter refused: true count: 0 why: gno.land/r/tests/vm/fvshare's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): v<VPBlock(1,0)>.N
// thrown writable, realm recovers refused: true count: 0 why: gno.land/r/tests/vm/fvshare's code was entered without crossing and is read-only (call its crossing function): v<VPBlock(1,0)>.N
// thrown writable, realm keeps refused: true count: 0 why: gno.land/r/tests/vm/fvshare's code was entered without crossing and is read-only (call its crossing function): v<VPBlock(1,0)>.N
// private refused: true count: 0 why: gno.land/r/tests/vm/fvshare's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): v<VPBlock(1,0)>.N
