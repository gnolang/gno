// PKGPATH: gno.land/r/tests/vm/recoverfvcross2
package recoverfvcross2

// Twin: a panic two crossing calls deep, recovered only outside the outer
// one, ends the run.

func try(fn func()) (rec any) {
	defer func() { rec = recover() }()
	fn()
	return
}

func outer(cur realm) { inner(cross(cur)) }
func inner(cur realm) { panic("crossed twice") }

func main(cur realm) {
	rec := try(func() { outer(cross(cur)) })
	println("not reached", rec)
}

// Error:
// crossed twice
