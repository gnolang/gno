// PKGPATH: gno.land/r/tests/vm/recoverfvcross
package recoverfvcross

// Twin: an explicit cross call that panics is not recovered across the
// realm boundary, even through a function value.
func boom(cur realm) { panic("crossed") }

func main(cur realm) {
	ops := []func(){func() { boom(cross(cur)) }}
	rec := func() (rec any) {
		defer func() { rec = recover() }()
		ops[0]()
		return
	}()
	println("not reached", rec)
}

// Error:
// crossed
