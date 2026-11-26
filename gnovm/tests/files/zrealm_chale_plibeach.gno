// PKGPATH: gno.land/r/tests/vm/chaleeach
package chaleeach

import lib "gno.land/p/demo/tests/chalelib"

var count int

// A literal written in place as an argument to a library-owned receiver's
// method: lent per the goal, but the method runs as the library's realm.
func main(cur realm) {
	lib.Default.Each(func(int) { count++ })
	println(count)
}

// Error:
// gno.land/r/tests/vm/chaleeach's code was entered without crossing and is read-only (call its crossing function): (const (ref(gno.land/r/tests/vm/chaleeach) package{})).count
