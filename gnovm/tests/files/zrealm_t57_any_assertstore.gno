// PKGPATH: gno.land/r/zrealm_t57_any_assertstore
package zrealm_t57_any_assertstore

var count int
var savedFn func()

// The callee asserts, stores the result, then calls it: an assertion not
// called at once escapes.
func assertStore(f any) {
	g := f.(func())
	savedFn = g
	g()
}

func main(cur realm) {
	println("refused:", try(func() { assertStore(func() { count++ }) }))
	println("later:", try(func() { savedFn() }))
	println("count:", count)
}

func try(fn func()) (r any) {
	defer func() { r = recover() }()
	fn()
	return nil
}

// Output:
// refused: gno.land/r/zrealm_t57_any_assertstore's code is read-only here: a function value writes only in the crossing call that made it (to write elsewhere, give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/zrealm_t57_any_assertstore) package{})).count
// later: gno.land/r/zrealm_t57_any_assertstore's code is read-only here: a function value writes only in the crossing call that made it (to write elsewhere, give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/zrealm_t57_any_assertstore) package{})).count
// count: 0
