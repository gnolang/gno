// PKGPATH: gno.land/r/test/sealifaceembed
package sealifaceembed

// Interface-embedding bypass: even when the legitimate concrete impl
// is unexported (e.g., grc20.fnTeller), a foreign realm can still
// satisfy a sealed interface by embedding the *interface itself* in
// a struct. The embedded interface value carries the package-
// qualified isSealed() forward; the struct's method set includes
// the promoted methods. Shadowing the methods the attacker cares
// about (e.g., Transfer that takes rlm) gives them full cur-leak
// despite the seal.
//
// This is the realistic Teller/Banker attack: an attacker realm
// obtains a legitimate seal.Sealed value from a public factory
// (e.g., Token.CallerTeller()), stores it in an embedded interface
// field, and shadows the rlm-typed methods.

import (
	"gno.land/p/test/seal"
)

// evilWrap embeds the sealed interface directly — no concrete type
// reference needed. If embedded-interface promotion satisfies the
// seal, this test compiles and the attack surface is real even when
// the legitimate concrete impl is unexported.
type evilWrap struct {
	seal.Sealed
}

// Shadow Hello() with attacker-controlled behavior.
func (e *evilWrap) Hello() string { return "evil via interface embedding" }

func main() {
	legit := seal.NewImpl("legit")
	evil := &evilWrap{Sealed: legit}
	var s seal.Sealed = evil
	println(s.Hello())
}

// Output:
// evil via interface embedding
