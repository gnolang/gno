// PKGPATH: gno.land/r/demo/curdeclimport
package curdeclimport

// An import name is a file-level binding (predefineRecursively2).

import cur "strings"

func main() {
	println(cur.ToUpper("a"))
}

// Error:
// gno.land/r/demo/curdeclimport/zrealm_cur_decl_import.gno:6:8-21: `cur` is a declaration-reserved identifier: it may only be declared as the first parameter of a crossing function
