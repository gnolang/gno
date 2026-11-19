// PKGPATH: gno.land/r/zrealm_entry_pfunc_anon
package zrealm_entry_pfunc_anon

import (
	"gno.land/p/demo/tests/launderattack/v0"
	"gno.land/r/tests/vm/pfuncvictim"
)

func main(cur realm) {
	pfuncvictim.ApplyAnon(cross(cur), struct{ F pfuncvictim.Fn }{F: launderattack.EvilWrite})
	println("victim:", pfuncvictim.Read())
}

// Error:
// cannot copy struct{F func(*gno.land/p/demo/tests/launderpkg/v0.Immutable)} outside realm gno.land/r/zrealm_entry_pfunc_anon (use a pointer)
