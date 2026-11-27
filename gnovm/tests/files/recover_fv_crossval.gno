// PKGPATH: gno.land/r/tests/vm/recoverfvcrossval
package recoverfvcrossval

// Twin: a stored crossing function called with cross panics; the panic
// crosses a realm boundary and is not recovered.
func boom(cur realm) { panic("crossed") }

var fns = []func(realm){boom}

func main(cur realm) {
	rec := func() (rec any) {
		defer func() { rec = recover() }()
		fns[0](cross(cur))
		return
	}()
	println("not reached", rec)
}

// Error:
// crossed
