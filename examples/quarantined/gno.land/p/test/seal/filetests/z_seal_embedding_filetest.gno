// PKGPATH: gno.land/r/test/sealembed
package sealembed

// Embedding-bypass test: if a foreign package can satisfy a sealed
// interface by *embedding* the legitimate in-package implementation,
// then the seal does not defend against the actual cur-leak threat
// model. The attacker would obtain a legitimate *seal.Impl, embed it
// in their own struct, and shadow whatever methods they want to
// control — including methods that take rlm in a real-world Teller
// or Banker interface.
//
// RESULT: Gno mirrors Go promotion semantics — the test compiles
// and runs. The seal is BYPASSABLE via embedding when the
// implementation type is exported. The attacker's Hello() runs
// while the embedded *seal.Impl supplies the promoted isSealed().
// See `// Output:` below for the attacker-controlled string.
//
// IMPLICATION for the cur-leak migration: sealing alone is NOT
// sufficient for Teller/Banker if their concrete impls are
// exported. `treasury.CoinsBanker` is exported (banker_coins.gno);
// `grc20.fnTeller` is unexported, which blocks concrete embedding
// but not interface-value embedding — see
// z_seal_iface_embedding_filetest.gno.

import (
	"gno.land/p/test/seal"
)

// foreignImpl embeds *seal.Impl. The promoted isSealed() method
// satisfies the seal's package-qualified requirement IF Gno follows
// Go's promotion rules.
type foreignImpl struct {
	*seal.Impl
}

// Shadow Hello() — this is the attack: foreign code controls what
// the "sealed" interface returns, while the seal check is satisfied
// via the embedded type's promoted isSealed().
func (f *foreignImpl) Hello() string { return "evil from r/test/sealembed" }

func main() {
	legit := seal.NewImpl("legit")
	evil := &foreignImpl{Impl: legit}
	var s seal.Sealed = evil
	println(s.Hello())
}

// Output:
// evil from r/test/sealembed
