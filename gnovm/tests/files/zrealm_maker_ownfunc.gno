// PKGPATH: gno.land/r/zrealm_maker_ownfunc
package zrealm_maker_ownfunc

// A realm hands a writer to a library that calls it back: its declared
// function passed as a value is refused; a literal written in place for a
// parameter Apply only calls writes while Apply runs.

import (
	"gno.land/p/demo/tests/launderpkg/v0"
	"gno.land/p/demo/tests/pfuncbind"
)

var obj = &launderpkg.Immutable{Field: "orig"}

func set(x *launderpkg.Immutable) { x.Field = "own" }

func main(cur realm) {
	println("declared:", try(func() { pfuncbind.Apply(set, obj) }), obj.Field)
	pfuncbind.Apply(func(x *launderpkg.Immutable) { x.Field = "lit" }, obj)
	println("literal:", obj.Field)
}

func try(fn func()) (r any) {
	defer func() { r = recover() }()
	fn()
	return nil
}

// Output:
// declared: gno.land/r/zrealm_maker_ownfunc's code is read-only here: a function value writes only in the crossing call that made it (to write elsewhere, give it a cur realm parameter and call it with cross): x<VPBlock(1,0)>.Field orig
// literal: lit
