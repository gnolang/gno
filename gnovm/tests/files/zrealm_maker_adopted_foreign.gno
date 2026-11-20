// PKGPATH: gno.land/r/zrealm_maker_adopted_foreign
package zrealm_maker_adopted_foreign

// A realm keeps closures another realm's helper built read-only: storing them
// is allowed, but they act for no realm, here or back in their maker's gate.

import "gno.land/r/tests/vm/cbvictim"

var (
	setX func(string)
	setY func([]string)
	buf  = []string{"orig"}
)

func Store(cur realm) { setX, setY = cbvictim.Writers() }

func CallBoth(cur realm) {
	try("x", func() { setX("y-gate") })
	try("y", func() { setY(buf) })
}

func try(name string, f func()) {
	defer func() {
		if r := recover(); r != nil {
			println(name, "refused:", r)
		}
	}()
	f()
}

func main(cur realm) {
	Store(cross(cur))
	CallBoth(cross(cur))
	println("owner:", cbvictim.Owner, "buf:", buf[0])
	cbvictim.RunS(cross(cur), setX)
	println("unreachable")
}

// Output:
// x refused: gno.land/r/tests/vm/cbvictim's code was entered without crossing and is read-only (call its crossing function): (const (ref(gno.land/r/tests/vm/cbvictim) package{})).Owner
// y refused: gno.land/r/tests/vm/cbvictim's code was entered without crossing and is read-only (call its crossing function): b<VPBlock(1,0)>[(const (0 int))]
// owner: victim buf: orig

// Error:
// gno.land/r/tests/vm/cbvictim's code is read-only here: a function value writes only in the crossing call that made it: (const (ref(gno.land/r/tests/vm/cbvictim) package{})).Owner
