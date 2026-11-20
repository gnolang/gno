// PKGPATH: gno.land/r/zrealm_t38_borrow_mv
// T38: a method value of this realm's own type bound in a borrowed /p/ frame
// writes in the making crossing call and is refused in a later one.

package zrealm_t38_borrow_mv

import (
	"gno.land/p/demo/tests/ifaceattack"
	"gno.land/r/tests/vm/t35c/victim"
)

type T struct{ Name string }

func (t *T) Set(k string, v any) bool { t.Name = v.(string); return true }

var Box = &T{Name: "orig"}

var kept func(string, any) bool

type picker struct{}

func (*picker) Pick() func() { return func() { kept("k", "later") } }

func main(cur realm) {
	f := ifaceattack.MakeSet(Box)
	f("k", "same")
	println("same:", Box.Name)
	kept = f
	victim.Keep(cross(cur), &picker{})
	victim.RunKept(cross(cur))
	println("later:", Box.Name)
}

// Output:
// same: same

// Error:
// gno.land/r/zrealm_t38_borrow_mv's code was entered without crossing and is read-only (call its crossing function): t<VPBlock(1,0)>.Name
