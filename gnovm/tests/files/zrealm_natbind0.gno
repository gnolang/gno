// PKGPATH: gno.land/r/test
package test

import (
	"std"
)

var node interface{}

func init() {
	node = std.GetHeight
}

func main() {
	// NOTE: this test uses GetHeight and GetChainID, which are "pure"
	// natively bound functions (ie. not indirections through a wrapper fn,
	// to convert the types to builtin go/gno identifiers).
	f := node.(func() int64)
	println(f())
	node = std.GetChainID
	g := node.(func() string)
	println(g())
}

// Output:
// 123
// dev

// Realm:
// switchrealm["gno.land/r/test"]
