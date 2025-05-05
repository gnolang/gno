// PKGPATH: gno.land/r/test
package test

import (
	"std"

	"gno.land/p/n2p5/loci"
)

var store *loci.LociStore

func init() {
	store = loci.New()
}

func main() {
	caller := std.PreviousRealm()

	store.Set([]byte("hello, world"))
	println(string(store.Get(caller.Address())))
}

// Output:
// hello, world
