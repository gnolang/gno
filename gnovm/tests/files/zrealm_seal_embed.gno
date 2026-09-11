// PKGPATH: gno.land/r/crossrealm
package crossrealm

// A struct that embeds a `realm` value inherits the hidden `.seal` marker
// through Go method promotion, and can override the identity methods
// (Address/PkgPath). Such a type must NOT satisfy `realm`: a promoted
// `.seal` no longer counts, so only a type that declares it directly (the
// runtime `.grealm`) satisfies the sealed interface. Otherwise identity
// (from the overrides) and liveness (from the embedded realm) could come
// from different concrete values. See gnovm/adr/realm_seal_embedding.md.
//
// The companion zrealm_seal_realm.gno covers the structural fake (no
// embedding); this one covers the embedding-promotion path.

type fooRealm struct {
	realm
	forged address
}

func (r fooRealm) Address() address { return r.forged }
func (r fooRealm) PkgPath() string  { return "gno.land/r/crossrealm" }

var _ realm = fooRealm{}

func main(cur realm) {}

// Error:
// gno.land/r/crossrealm/zrealm_seal_embed.gno:23:5-25: gno.land/r/crossrealm.fooRealm does not implement .uverse.realm (missing method .seal)
