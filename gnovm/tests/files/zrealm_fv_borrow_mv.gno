// PKGPATH: gno.land/r/zrealm_fv_borrow_mv
// A method value of this realm's own type bound in a borrowed /p/ frame
// never writes: refused in the making crossing call and in a later one.

package zrealm_fv_borrow_mv

import (
	"gno.land/p/demo/tests/ifaceattack"
	"gno.land/r/tests/vm/pickvault/victim"
)

type T struct{ Name string }

func (t *T) Set(k string, v any) bool { t.Name = v.(string); return true }

var Box = &T{Name: "orig"}

var kept func(string, any) bool

type picker struct{}

func (*picker) Pick() func() { return func() { kept("k", "later") } }

func main(cur realm) {
	f := ifaceattack.MakeSet(Box)
	println("same:", try(func() { f("k", "same") }), Box.Name)
	kept = f
	victim.Keep(cross(cur), &picker{})
	victim.RunKept(cross(cur))
	println("later:", Box.Name)
}

func try(fn func()) (r any) {
	defer func() { r = recover() }()
	fn()
	return nil
}

// Output:
// same: gno.land/r/zrealm_fv_borrow_mv's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): t<VPBlock(1,0)>.Name orig

// Error:
// gno.land/r/zrealm_fv_borrow_mv's code was entered without crossing and is read-only (call its crossing function): t<VPBlock(1,0)>.Name
