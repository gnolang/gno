// PKGPATH: gno.land/r/tests/vm/recoverfvrealm
package recoverfvrealm

import "gno.land/p/demo/tests/recoverfv"

var (
	c   = recoverfv.New()
	ops = []func(){
		func() { panic("own") },
		func() { c.Set("x") },
	}
)

// A realm's stored function value panics; the realm's own caller recovers.
func main(cur realm) {
	for i := range ops {
		println(i, try(ops[i]))
	}
	println("OUTCOME caught")
}

func try(fn func()) (rec any) {
	defer func() { rec = recover() }()
	fn()
	return
}

// Output:
// 0 own
// 1 int is not of type string
// OUTCOME caught
