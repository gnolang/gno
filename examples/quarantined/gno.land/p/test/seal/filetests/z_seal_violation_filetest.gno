// PKGPATH: gno.land/r/test/sealviolation
package sealviolation

// Regression test: verify that an interface declared in another
// package with an unexported method cannot be satisfied by a type
// declared in this package. If this filetest compiles, Gno's
// sealed-interface semantics differ from Go's and the cur-leak
// mitigation strategy (sealing /p/ interfaces like grc20.Teller via
// an unexported method) does NOT work.
//
// The expected behavior is a preprocess/typecheck error at the
// assignment line: foreignImpl's isSealed() method belongs to
// gno.land/r/test/sealviolation, not gno.land/p/test/seal, so it
// doesn't satisfy seal.Sealed's seal.isSealed() requirement.

import (
	"gno.land/p/test/seal"
)

type foreignImpl struct{}

func (f *foreignImpl) Hello() string { return "evil" }

// This method is named isSealed() but it belongs to package
// sealviolation, not package seal. Structural interface matching
// should look for *seal.isSealed* on the type, not just any method
// named isSealed.
func (f *foreignImpl) isSealed() {}

func main() {
	var s seal.Sealed = &foreignImpl{} // should be a compile-time error
	println(s.Hello())
}

// Error:
// gno.land/r/test/sealviolation/z_seal_violation.gno:31:6-36: *gno.land/r/test/sealviolation.foreignImpl does not implement gno.land/p/test/seal.Sealed (missing method isSealed)
