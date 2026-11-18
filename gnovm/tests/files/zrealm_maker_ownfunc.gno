// PKGPATH: gno.land/r/zrealm_maker_ownfunc
package zrealm_maker_ownfunc

// A realm hands its own declared writer to a library that calls it back:
// the realm's own function writes for it, as before.

import (
	"gno.land/p/demo/tests/launderpkg/v0"
	"gno.land/p/demo/tests/pfuncbind"
)

var obj = &launderpkg.Immutable{Field: "orig"}

func set(x *launderpkg.Immutable) { x.Field = "own" }

func main(cur realm) {
	pfuncbind.Apply(set, obj)
	println("field:", obj.Field)
}

// Output:
// field: own
