// PKGPATH: gno.land/r/zrealm_t57_any_passkeep
package zrealm_t57_any_passkeep

var count int
var saved any

// The callee passes the `any`, or its switched value, on to a function that
// keeps it, then calls what was kept: the loan ends at the keeper.
func passKeep(f any) {
	keep(f)
	saved.(func())()
}

func switchKeep(f any) {
	switch g := f.(type) {
	case func():
		keep(g)
		g()
	}
}

func keep(f any) { saved = f }

func main(cur realm) {
	println("pass:", try(func() { passKeep(func() { count++ }) }))
	println("switch:", try(func() { switchKeep(func() { count++ }) }))
	println("later:", try(func() { saved.(func())() }))
	println("count:", count)
}

func try(fn func()) (r any) {
	defer func() { r = recover() }()
	fn()
	return nil
}

// Output:
// pass: gno.land/r/zrealm_t57_any_passkeep's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/zrealm_t57_any_passkeep) package{})).count
// switch: gno.land/r/zrealm_t57_any_passkeep's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/zrealm_t57_any_passkeep) package{})).count
// later: gno.land/r/zrealm_t57_any_passkeep's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/zrealm_t57_any_passkeep) package{})).count
// count: 0
