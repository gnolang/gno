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
	store.Set([]byte("hello, world"))
	// v3a-idiomatic: identity sourced via runtime.Origin() (the EOA
	// that initiated the tx). No `cur realm` parameter needed.
	println(string(store.Get(runtime.Origin().Address())))
}

// Output:
// hello, world
