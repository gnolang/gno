// Package seal documents whether Gno's sealed-interface pattern
// (unexported marker method on an interface) provides a security
// boundary. CONCLUSION: it does NOT. See the filetests for proof.
//
// Search terms: sealed interface, embedding bypass, unexported marker
// method, capability-token leak, isSealed bypass, cur leak via
// embedding, rlm leak via interface, defense in depth, canonical-impl
// allowlist, IsCanonical pattern, allowancesender, Teller seal,
// Banker seal.
//
// Filetests (read these before reaching for the sealed-marker
// pattern as a security control):
//   - z_seal_violation_filetest.gno — direct redeclaration of the
//     unexported marker in a foreign package FAILS. (The original
//     intended property survives.)
//   - z_seal_embedding_filetest.gno — concrete-type embedding
//     BYPASSES the seal when the legitimate impl is exported. The
//     promoted unexported marker retains its origin package's
//     identity and satisfies the interface.
//   - z_seal_iface_embedding_filetest.gno — embedding the interface
//     value itself BYPASSES the seal regardless of whether the
//     legitimate concrete impl is exported. This is the realistic
//     Teller/Banker attack: hold a legit Teller obtained from
//     Token.CallerTeller(), embed it, shadow Transfer to capture rlm.
//
// Implication for the cur-leak migration: don't seal as a defense.
// Use Reader/Writer split, Authority-pattern with private concrete,
// or canonical-impl allowlist (see migration guide).
package seal

// Sealed is intended to be implementable only from within the seal
// package. The unexported isSealed() method is the seal — any foreign
// type that declares an `isSealed()` method declares
// `r/foreign.isSealed`, not `p/test/seal.isSealed`, so structural
// satisfaction fails.
type Sealed interface {
	Hello() string
	isSealed()
}

// Impl is the legitimate in-package implementation of Sealed.
type Impl struct {
	Name string
}

func NewImpl(name string) *Impl {
	return &Impl{Name: name}
}

func (i *Impl) Hello() string { return "hello, " + i.Name }
func (i *Impl) isSealed()     {}

// Accept demonstrates a consumer that takes a Sealed value. In real
// code this is the surface a cur-leaking interface would expose; the
// security property of sealing is that Accept can only ever receive a
// package-internal *Impl, not an attacker-supplied type.
func Accept(s Sealed) string {
	return s.Hello()
}
