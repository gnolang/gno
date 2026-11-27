// PKGPATH: gno.land/r/tests/vm/recoverfvlent
package recoverfvlent

import "gno.land/p/demo/tests/recoverfv"

// A lent literal panics inside a library; the realm that wrote it recovers.
func main(cur realm) {
	c := recoverfv.New()
	rec := func() (rec any) {
		defer func() { rec = recover() }()
		recoverfv.Each([]int{1}, func(int) { panic("lent") })
		return
	}()
	println("lent:", rec)
	rec = func() (rec any) {
		defer func() { rec = recover() }()
		recoverfv.Each([]int{1}, func(int) { c.Set("x") })
		return
	}()
	println("lent method:", rec)
	println("OUTCOME caught")
}

// Output:
// lent: lent
// lent method: int is not of type string
// OUTCOME caught
