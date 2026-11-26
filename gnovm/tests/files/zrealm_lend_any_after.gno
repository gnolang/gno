// PKGPATH: gno.land/r/zrealm_lend_any_after
package zrealm_lend_any_after

var count int

// The callee returns the literal; it is called after the callee returned.
func ret(f any) func() { return f.(func()) }

func retSwitch(f any) func() {
	switch g := f.(type) {
	case func():
		return g
	}
	return nil
}

func main(cur realm) {
	println("assert:", try(func() { ret(func() { count++ })() }))
	println("switch:", try(func() { retSwitch(func() { count++ })() }))
	println("count:", count)
}

func try(fn func()) (r any) {
	defer func() { r = recover() }()
	fn()
	return nil
}

// Output:
// assert: gno.land/r/zrealm_lend_any_after's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/zrealm_lend_any_after) package{})).count
// switch: gno.land/r/zrealm_lend_any_after's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/zrealm_lend_any_after) package{})).count
// count: 0
