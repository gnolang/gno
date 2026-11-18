// PKGPATH: gno.land/r/zrealm_maker_methodexpr
package zrealm_maker_methodexpr

// Another realm's method expression used as a value is refused at preprocess.

import "gno.land/r/tests/vm/cbvictim"

func main(cur realm) {
	f := (*cbvictim.T).SetName
	f(cbvictim.Box, "picked")
	println("name:", cbvictim.Box.Name)
}

// Error:
// gno.land/r/zrealm_maker_methodexpr/zrealm_maker_methodexpr.gno:9:7-28: cannot use non-crossing function gno.land/r/tests/vm/cbvictim.T.SetName of another realm as a value (call it, or wrap it in a closure)
