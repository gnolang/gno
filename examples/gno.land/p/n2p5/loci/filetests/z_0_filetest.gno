// PKGPATH: gno.land/r/test
package test

import (
	"chain/runtime"

	"gno.land/p/n2p5/loci"
)

var store *loci.LociStore

func init() {
	store = loci.New()
}

func main() {
	caller := runtime.PreviousRealm()

	store.Set([]byte("hello, world"))
	println(string(store.Get(caller.Address())))
}

// Output:
// hello, world
