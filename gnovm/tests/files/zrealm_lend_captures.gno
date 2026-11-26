// PKGPATH: gno.land/r/zrealm_lend_captures
package zrealm_lend_captures

var count int
var kept func()

// outer's f is lent; the literal capturing it is lent to mid, which hands it
// to keep, which stores it: the literal and f both lose their loans, so f is
// refused in outer and the kept literal runs read-only.
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
// after escape: gno.land/r/zrealm_lend_captures's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/zrealm_lend_captures) package{})).count
// kept: gno.land/r/zrealm_lend_captures's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/zrealm_lend_captures) package{})).count
// count: 0
