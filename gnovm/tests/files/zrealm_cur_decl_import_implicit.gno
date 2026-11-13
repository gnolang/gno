// PKGPATH: gno.land/r/demo/curdeclimportimplicit
package curdeclimportimplicit

// An import with no alias binds the package's own name, here `cur`, so
// `cur.Previous().IsUser()` would call whatever that package defines. go/types
// records the binding in Implicits, not Defs, and the gate reads both.

import "filetests/extern/cur"

func Guarded(_ realm) string {
	if !cur.Previous().IsUser() {
		panic("unauthorized")
	}
	return "admin action ran"
}

func main(cur realm) {
	println(Guarded(cur))
}

// Output:
// admin action ran

// TypeCheckError:
// gno.land/r/demo/curdeclimportimplicit/zrealm_cur_decl_import_implicit.gno:8:8: `cur` is a contextual keyword: it may only be declared as the first parameter of a crossing function
