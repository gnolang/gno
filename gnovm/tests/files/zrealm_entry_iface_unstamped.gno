// PKGPATH: gno.land/r/zrealm_entry_iface_unstamped
// A method value of the victim's own avl tree, built by a caller's library while
// view-only, has no stamp; the victim imports avl, yet the call acts for no realm.
package zrealm_entry_iface_unstamped

import (
	"gno.land/p/demo/tests/ifaceattack"
	"gno.land/r/tests/vm/ifacevictim"
)

func main(cur realm) {
	mk := ifaceattack.MakeSet
	ifacevictim.ApplySet(cross(cur), mk(ifacevictim.Tree()))
	println("victim has k:", ifacevictim.TreeHas())
}

// Error:
// cannot modify object owned by another realm from library code (call a function of the owning realm): tree<VPBlock(1,0)>.node
