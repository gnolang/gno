// PKGPATH: gno.land/r/zrealm_lend_any_other
package zrealm_lend_any_other

import "gno.land/r/tests/vm/lendanyother"

var count int

// Another realm's callee runs read-only: the literal it calls cannot write.
func main(cur realm) {
	r := try(func() { lendanyother.RunAny(func() { count++ }) })
	println("refused:", r)
	println("count:", count)
}

func try(fn func()) (r any) {
	defer func() { r = recover() }()
	fn()
	return nil
}

// Output:
// refused: gno.land/r/zrealm_lend_any_other's code was entered without crossing and is read-only (call its crossing function): (const (ref(gno.land/r/zrealm_lend_any_other) package{})).count
// count: 0
