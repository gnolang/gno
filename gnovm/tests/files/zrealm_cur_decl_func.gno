// PKGPATH: gno.land/r/demo/curdeclfunc
package curdeclfunc

// A package-level func name is a binding (predefineRecursively2), so a
// function named `cur` is refused like any other declaration of the name.

func cur() int { return 1 }

func main() {
	println(cur())
}

// Error:
// gno.land/r/demo/curdeclfunc/zrealm_cur_decl_func.gno:7:1-28: `cur` is reserved: it may only be declared as the first parameter of a crossing function
