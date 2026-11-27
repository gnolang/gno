// PKGPATH: gno.land/r/zrealm_fvshare_chal1
// A writer thrown into a stranger's deferred catcher never writes, however
// the stranger keeps it; Private, a value called from a local slice, neither.
package zrealm_fvshare_chal1

import "gno.land/r/tests/vm/fvshare"

type K struct {
	f    func()
	c    *fvshare.Ctx
	call bool
}

func (k *K) Show(c *fvshare.Ctx) { k.c = c }

func (k *K) Catch() {
	v := recover()
	println("stranger got:", v != nil)
	if f, ok := v.(func()); ok {
		if k.call {
			f()
		}
		if k.c != nil {
			k.c.Cb = f
		}
		k.f = f
	}
}

func (k *K) Pick() func() { return k.f }

func try(name string, f func()) {
	r := revive(f)
	println(name, "refused:", r != nil, "count:", fvshare.Count(), "why:", r)
}

func main(cur realm) {
	try("pick", func() { fvshare.ChalPick(cross(cur), &K{}) })
	try("catch calls", func() { fvshare.ChalPick(cross(cur), &K{call: true}) })
	try("slot", func() { fvshare.ChalSlot(cross(cur), &K{}) })
	try("private", func() { fvshare.Private(cross(cur)) })
}

// Output:
// stranger got: true
// pick refused: true count: 0 why: gno.land/r/zrealm_fvshare_chal1's code was entered without crossing and is read-only (call its crossing function): k<VPBlock(2,0)>.f
// stranger got: true
// catch calls refused: true count: 0 why: gno.land/r/tests/vm/fvshare's code was entered without crossing and is read-only (call its crossing function): v<VPBlock(1,0)>.N
// slot refused: true count: 0 why: gno.land/r/zrealm_fvshare_chal1's code was entered without crossing and is read-only (call its crossing function): k<VPBlock(1,0)>.c
// private refused: true count: 0 why: gno.land/r/tests/vm/fvshare's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): v<VPBlock(1,0)>.N
