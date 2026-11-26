// PKGPATH: gno.land/r/zrealm_t57_captures
package zrealm_t57_captures

var count int
var kept func()

// outer's f is lent; the literal capturing it is lent to mid, which hands it
// to keep: the literal loses its loan, f keeps its own until outer returns,
// and the kept literal then runs read-only.
func outer(f func()) {
	mid(func() { f() })
	f()
}

func mid(g func()) { keep(g) }

func keep(g func()) { kept = g }

func main(cur realm) {
	println("after escape:", try(func() { outer(func() { count++ }) }))
	println("kept:", try(func() { kept() }))
	println("count:", count)
}

func try(fn func()) (r any) {
	defer func() { r = recover() }()
	fn()
	return nil
}

// Output:
// after escape: undefined
// kept: gno.land/r/zrealm_t57_captures's code is read-only here: a function value writes only in the crossing call that made it (to write elsewhere, give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/zrealm_t57_captures) package{})).count
// count: 1
