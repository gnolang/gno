// PKGPATH: gno.land/r/demo/curdeclintparam
package curdeclintparam

// A parameter named `cur` of a non-realm type is a declaration of the
// reserved name outside the one site that may bind it.

func f(cur int) int { return cur }

func main() {
	println(f(1))
	println("ok")
}

// Output:
// 1
// ok

// TypeCheckError:
// gno.land/r/demo/curdeclintparam/zrealm_cur_decl_int_param.gno:7:8: `cur` is a contextual keyword: it may only be declared as the first parameter of a crossing function
