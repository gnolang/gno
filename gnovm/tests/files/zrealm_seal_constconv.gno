// PKGPATH: gno.land/r/crossrealm
package crossrealm

// A typed constant conversion to realm must enforce the hidden seal just like
// every other interface conversion. Without that check, the conversion's
// static type becomes realm while the stored value keeps the fake type, so
// calls dispatch to these attacker-controlled methods.
type fake int

func (fake) Address() address  { return address("") }
func (fake) PkgPath() string   { return "gno.land/r/crossrealm" }
func (fake) Previous() realm   { var z realm; return z }
func (fake) IsCode() bool      { return true }
func (fake) IsUser() bool      { return false }
func (fake) IsUserCall() bool  { return false }
func (fake) IsUserRun() bool   { return false }
func (fake) IsEphemeral() bool { return false }
func (fake) IsCurrent() bool   { return true }
func (fake) Sub(string) realm  { var z realm; return z }
func (fake) Subpath() string   { return "" }
func (fake) String() string    { return "forged" }

type stringer interface{ String() string }

func main(cur realm) {
	_ = stringer(fake(1))
	_ = realm(fake(1))
}

// Error:
// gno.land/r/crossrealm/zrealm_seal_constconv.gno:27:6-20: gno.land/r/crossrealm.fake does not implement .uverse.realm (missing method .seal)

// TypeCheckError:
// gno.land/r/crossrealm/zrealm_seal_constconv.gno:27:12: cannot convert fake(1) (constant 1 of int type fake) to type realm: fake does not implement gno0p9.realm (missing method seal)
