// PKGPATH: gno.land/r/demo/reflectid
package reflectid

import (
	"chain/reflect"
)

type thing struct{ name string }

var (
	fromInit *thing
	idInInit reflect.ObjectID
	okInInit bool
)

func init() {
	fromInit = &thing{name: "a"}
	// init has not returned across a realm boundary yet, so the VM has not
	// persisted fromInit and has no ID for it.
	idInInit, okInInit = reflect.ObjectIDOf(fromInit)
}

func main(cur realm) {
	println("stamped during init:", okInInit)
	println("zero ID during init:", idInInit.IsZero())
	// It was an object all along; only the ID was missing. That is the
	// distinction a caller needs to tell "ask again later" from "never".
	println("was an object during init:", reflect.HasIdentity(fromInit))

	// init returned, so the realm persisted fromInit and stamped it.
	id, ok := reflect.ObjectIDOf(fromInit)
	println("stamped after init:", ok)
	println("zero ID after init:", id.IsZero())

	// Reading twice agrees: this reads an identity, it does not issue one.
	again, _ := reflect.ObjectIDOf(fromInit)
	println("stable across reads:", id == again)

	// A second object created now is in the same window fromInit was in.
	fresh := &thing{name: "b"}
	_, freshOK := reflect.ObjectIDOf(fresh)
	println("fresh object stamped:", freshOK)
	println("fresh object is an object:", reflect.HasIdentity(fresh))

	println(".")
}

// Output:
// stamped during init: false
// zero ID during init: true
// was an object during init: true
// stamped after init: true
// zero ID after init: false
// stable across reads: true
// fresh object stamped: false
// fresh object is an object: true
// .
