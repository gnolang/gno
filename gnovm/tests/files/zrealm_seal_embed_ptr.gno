// PKGPATH: gno.land/r/crossrealm
package crossrealm

// Companion to zrealm_seal_embed.gno: here the hidden `.seal` marker is
// promoted through a POINTER to an embedded struct (fooRealm embeds *inner,
// inner embeds a realm). This reaches the marker via a pointer-deref field hop
// (VPDerefField) rather than the plain value field (VPField) of
// zrealm_seal_embed.gno, and must be rejected the same way: only a method
// declared directly on the concrete type (the runtime `.grealm`) satisfies the
// sealed `realm` interface. See gnovm/adr/realm_seal_embedding.md.

type inner struct {
	realm
}

type fooRealm struct {
	*inner
	forged address
}

func (r fooRealm) Address() address { return r.forged }
func (r fooRealm) PkgPath() string  { return "gno.land/r/crossrealm" }

var _ realm = fooRealm{}

func main(cur realm) {}

// Error:
// gno.land/r/crossrealm/zrealm_seal_embed_ptr.gno:24:5-25: gno.land/r/crossrealm.fooRealm does not implement .uverse.realm (missing method .seal)
