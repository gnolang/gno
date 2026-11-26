// PKGPATH: gno.land/r/zrealm_fvshare_defer
// A writer handed over by a deferred call never writes; Private, a value
// called from a local slice, neither.
package zrealm_fvshare_defer

import "gno.land/r/tests/vm/fvshare"

type D struct{}

func (d *D) Use(v any, set func(any)) { set(v) }

// GK keeps a deferred call's argument behind a getter closure and hands it
// back by a return or a panic.
type GK struct {
	put func(g func() func())
	get func() func() func()
}

func (k *GK) Keep(f func())  { k.put(func() func() { return f }) }
func (k *GK) Throw(f func()) { panic(func() func() { return f }) }

func (k *GK) Pick() func() {
	g := k.get()
	if g == nil {
		return nil
	}
	return g()
}

func newGK() *GK {
	var stash func() func()
	return &GK{
		put: func(g func() func()) { stash = g },
		get: func() func() func() { return stash },
	}
}

var took any

func other(cur realm, v any) {
	took = v
	if f, ok := v.(func()); ok {
		f()
	}
}

func try(name string, f func()) {
	r := revive(f)
	println(name, "refused:", r != nil, "count:", fvshare.Count(), "why:", r)
}

func main(cur realm) {
	h := &D{}
	for _, shape := range []string{"direct", "struct", "slice", "method"} {
		shape := shape
		try("deferred "+shape, func() { fvshare.DeferArgs(cross(cur), h, other, shape) })
	}
	try("deferred crossing", func() { fvshare.DeferArgs(cross(cur), h, other, "crossing") })
	try("crossing kept", func() { fvshare.CallAny(cross(cur), took) })
	try("deferred keep, picked", func() { fvshare.DeferKeep(cross(cur), newGK()) })
	try("deferred throw, getter", func() { fvshare.DeferThrow(cross(cur), newGK()) })
	try("private", func() { fvshare.Private(cross(cur)) })
}

// Output:
// deferred direct refused: true count: 0 why: gno.land/r/tests/vm/fvshare's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): v<VPBlock(1,0)>.N
// deferred struct refused: true count: 0 why: gno.land/r/tests/vm/fvshare's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): v<VPBlock(1,0)>.N
// deferred slice refused: true count: 0 why: gno.land/r/tests/vm/fvshare's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): v<VPBlock(1,0)>.N
// deferred method refused: true count: 0 why: gno.land/r/tests/vm/fvshare's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): v<VPBlock(1,0)>.N
// deferred crossing refused: true count: 0 why: gno.land/r/tests/vm/fvshare's code was entered without crossing and is read-only (call its crossing function): v<VPBlock(1,0)>.N
// crossing kept refused: true count: 0 why: gno.land/r/tests/vm/fvshare's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): v<VPBlock(1,0)>.N
// deferred keep, picked refused: true count: 0 why: gno.land/r/tests/vm/fvshare's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): v<VPBlock(1,0)>.N
// deferred throw, getter refused: true count: 0 why: gno.land/r/tests/vm/fvshare's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): v<VPBlock(1,0)>.N
// private refused: true count: 0 why: gno.land/r/tests/vm/fvshare's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): v<VPBlock(1,0)>.N
