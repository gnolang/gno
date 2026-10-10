// PKGPATH: gno.land/r/demo/reflectid
package reflectid

// What the VM knows about an object, printed rather than asserted.
//
// Every value below is deterministic: run this on any machine, at any height,
// and the same addresses come out. That is the whole point. Nothing here is a
// hash of wall-clock time, a counter the realm keeps, or anything the realm
// chose at all.

import (
	"chain/reflect"
)

type Account struct{ label string }

var first, second *Account

func init() {
	first = &Account{label: "alice"}
	second = &Account{label: "bob"}
}

func main(cur realm) {
	a, _ := reflect.Of(first)
	b, _ := reflect.Of(second)

	// The human-readable rendering: which type, which one of them, and where
	// to find it.
	println("first  ", a.String())
	println("second ", b.String())
	println("")

	// Broken out. The address is what goes in an event, a registry or a banker
	// call; everything else is provenance.
	println("address    ", a.Address())
	println("id         ", a.ID())
	println("created by ", a.PkgPath())
	println("type       ", a.Type())
	println("")

	// Two objects of one type in one realm. The realm chose neither address and
	// could not have made them collide.
	println("distinct addresses:", a.Address() != b.Address())
	println("same type:         ", a.Type() == b.Type())
	println("same realm:        ", a.PkgPath() == b.PkgPath())
	println("")

	// Reading is not issuing. Asking twice gives the same answer, and asking
	// persists nothing and advances no clock.
	again, _ := reflect.Of(first)
	println("stable across reads:", a == again)
	println("")

	// An object created by this call is inside the stamping window: the VM has
	// not persisted it yet, so there is no identity and no address to derive.
	// It is still an object, which is what tells "not yet" from "never".
	fresh := &Account{label: "carol"}
	_, ok := reflect.Of(fresh)
	println("fresh object has an address:", ok)
	println("fresh object is an object:  ", reflect.HasIdentity(fresh))
	println("a scalar is an object:      ", reflect.HasIdentity(42))

	println(".")
}

// Output:
// first   gno.land/r/demo/reflectid.Account#8 (g1delmyrysc3x2xdq92j2ttmykkfj5rs6r682dw0)
// second  gno.land/r/demo/reflectid.Account#10 (g1y6em47cffhtec6cg0a60q60gww5twxg2plgwkv)
//
// address     g1delmyrysc3x2xdq92j2ttmykkfj5rs6r682dw0
// id          0f0fab98e0c5e59ebb5c1cae1ba307b5adcb6026:8
// created by  gno.land/r/demo/reflectid
// type        gno.land/r/demo/reflectid.Account
//
// distinct addresses: true
// same type:          true
// same realm:         true
//
// stable across reads: true
//
// fresh object has an address: false
// fresh object is an object:   true
// a scalar is an object:       false
// .
