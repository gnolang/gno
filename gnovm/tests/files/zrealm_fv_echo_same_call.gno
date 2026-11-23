// PKGPATH: gno.land/r/zrealm_fv_echo_same_call
package zrealm_fv_echo_same_call

// R2b: fvecho's writer is handed back by another realm's code inside the
// crossing call that made it; the copy has no maker, so R2a refuses it (the
// error names R2a although the making call is still running).
import "gno.land/r/tests/vm/fvecho"

func main(cur realm) {
	fvecho.Echo(cross(cur))
	println("owner:", fvecho.Owner)
}

// Error:
// gno.land/r/tests/vm/fvecho's code is read-only here: a function value writes only in the crossing call that made it: (const (ref(gno.land/r/tests/vm/fvecho) package{})).Owner
