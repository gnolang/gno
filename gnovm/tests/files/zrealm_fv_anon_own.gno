// PKGPATH: gno.land/r/fvanonown
package fvanonown

import "gno.land/r/tests/vm/fvx"

func main(cur realm) {
	x := fvx.Stats()
	x.A = 9
	println(x.A)
	a := fvx.StatsArr()
	a[1].A = 7
	println(a[1].A)
}

// Output:
// 9
// 7
