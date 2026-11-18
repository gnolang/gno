// PKGPATH: gno.land/r/valueban
package valueban

import "gno.land/r/tests/vm/ctw"

func main(cur realm) {
	p := ctw.NewLocal(1)
	println(p.N, ctw.NewLocal(2).N)
	c := *p
	println(c.N)
}

// Output:
// 1 2
// 1

// TypeCheckError:
// gno.land/r/valueban/zrealm_value_ban.gno:9:2: cannot hold a value of gno.land/r/tests/vm/ctw.Local outside its realm (use a pointer); gno.land/r/valueban/zrealm_value_ban.gno:9:7: cannot hold a value of gno.land/r/tests/vm/ctw.Local outside its realm (use a pointer)
