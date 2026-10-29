// PKGPATH: gno.land/r/demo/curdeclimport
package curdeclimport

// An import name is a file-level binding (predefineRecursively2).

import cur "strings"

func main() {
	println(cur.ToUpper("a"))
}

// Output:
// A

// TypeCheckError:
// gno.land/r/demo/curdeclimport/zrealm_cur_decl_import.gno:6:8: `cur` is a contextual keyword: it may only be declared as the first parameter of a crossing function
