// PKGPATH: gno.land/r/zrealm_t53_recv_redecl
package zrealm_t53_recv_redecl

var count int

type F func()

var stored F = func() { count += 100 }

// The receiver g is reused by `:=` (gno refuses a bare literal here: Define2
// cannot change F to func()); the first call still runs the stored value.
func (g F) Run() {
	g()
	g, ok := F(func() {}), true
	_ = ok
}

func main(cur realm) {
	println("before:", count)
	stored.Run()
	println("after:", count)
}

// Output:
// before: 0

// Error:
// gno.land/r/zrealm_t53_recv_redecl's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/zrealm_t53_recv_redecl) package{})).count
