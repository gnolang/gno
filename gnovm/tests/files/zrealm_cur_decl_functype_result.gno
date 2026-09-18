// PKGPATH: gno.land/r/demo/curdeclfunctyperes
package curdeclfunctyperes

// A function type's result named `cur` is refused by name.

type F func() (cur int)

func main() {
	var f F
	_ = f
}

// Error:
// gno.land/r/demo/curdeclfunctyperes/zrealm_cur_decl_functype_result.gno:6:8-24: `cur` is reserved: it may only be declared as the first parameter of a crossing function
