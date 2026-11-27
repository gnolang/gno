// PKGPATH: gno.land/r/demo/nomutable
package nomutable

// mutable(x) was the owner's write grant; it is gone, so the name no
// longer resolves.

type T struct{ n int }

var t = &T{}

func main(cur realm) {
	println(mutable(t))
}

// Error:
// gno.land/r/demo/nomutable/zrealm_entry_no_mutable.gno:12:10-17: name mutable not declared

// TypeCheckError:
// gno.land/r/demo/nomutable/zrealm_entry_no_mutable.gno:12:10: undefined: mutable
