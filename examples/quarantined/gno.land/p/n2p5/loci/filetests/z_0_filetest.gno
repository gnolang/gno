// PKGPATH: gno.land/r/test
package test

import (
	"gno.land/p/n2p5/loci"
)

var store *loci.LociStore

func init() {
	store = loci.New()
}

func main(cur realm) {
	store.Set(0, cur, []byte("hello, world"))
	println(string(store.Get(cur.Previous().Address())))
}

// Output:
// hello, world
