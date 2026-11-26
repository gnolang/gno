// PKGPATH: gno.land/r/zrealm_lend_box_shadow
package zrealm_lend_box_shadow

import (
	"gno.land/r/tests/vm/lendboxattack"
	"gno.land/r/tests/vm/lendboxvictim"
)

func main(cur realm) {
	println("before:", lendboxvictim.Total())
	lendboxattack.Hook(cross(cur))
	println("refused:", revive(func() { lendboxvictim.RunHook(cross(cur)) }))
	println("after:", lendboxvictim.Total())
}

// Output:
// before: 0
// refused: gno.land/r/tests/vm/lendboxvictim's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): v<VPBlock(1,0)>.N
// after: 0
