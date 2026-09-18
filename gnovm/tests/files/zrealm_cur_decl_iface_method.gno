// PKGPATH: gno.land/r/demo/curdecliface
package curdecliface

// An interface method's later parameter named `cur` is refused by name.

type I interface{ M(x int, cur int) }

func main() {
	var i I
	_ = i
}

// Error:
// gno.land/r/demo/curdecliface/zrealm_cur_decl_iface_method.gno:6:20-36: `cur` is reserved: it may only be declared as the first parameter of a crossing function
