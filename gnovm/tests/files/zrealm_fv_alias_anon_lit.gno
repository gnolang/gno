// PKGPATH: gno.land/r/fvaliasanonlit
package fvaliasanonlit

import "gno.land/r/tests/vm/fvx"

func main(cur realm) {
	var a fvx.S
	b := struct{ N int }{1}
	c := fvx.S{N: 2}
	d := new(fvx.S)
	println(a.N, b.N, c.N, d.N)
}

// Output:
// 0 1 2 0
