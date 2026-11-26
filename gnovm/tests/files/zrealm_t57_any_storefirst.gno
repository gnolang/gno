// PKGPATH: gno.land/r/zrealm_t57_any_storefirst
package zrealm_t57_any_storefirst

var count int
var saved any

// The callee stores the `any` before switching on it: f escapes, no loan.
func storeFirst(f any) {
	saved = f
	switch g := f.(type) {
	case func():
		g()
	}
}

func main(cur realm) {
	println("refused:", try(func() { storeFirst(func() { count++ }) }))
	println("later:", try(func() { saved.(func())() }))
	println("count:", count)
}

func try(fn func()) (r any) {
	defer func() { r = recover() }()
	fn()
	return nil
}

// Output:
// refused: gno.land/r/zrealm_t57_any_storefirst's code is read-only here: a function value writes only in the crossing call that made it (to write elsewhere, give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/zrealm_t57_any_storefirst) package{})).count
// later: gno.land/r/zrealm_t57_any_storefirst's code is read-only here: a function value writes only in the crossing call that made it (to write elsewhere, give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/zrealm_t57_any_storefirst) package{})).count
// count: 0
