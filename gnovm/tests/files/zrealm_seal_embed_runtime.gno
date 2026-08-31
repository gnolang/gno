// PKGPATH: gno.land/r/crossrealm
package crossrealm

// The seal must also hold at the runtime satisfaction paths, not only static
// assignment. A forged embedder is boxed in an interface{} and asserted to
// `realm` with the comma-ok form: because a promoted `.seal` does not satisfy
// the sealed interface, ok is false — no forged realm escapes and no panic.
// See gnovm/adr/realm_seal_embedding.md.

type fooRealm struct {
	realm
	forged address
}

func (r fooRealm) Address() address { return r.forged }
func (r fooRealm) PkgPath() string  { return "gno.land/r/crossrealm" }

func main(cur realm) {
	var box any = fooRealm{realm: cur}
	_, ok := box.(realm)
	println("assertion ok:", ok)
}

// Output:
// assertion ok: false
