// PKGPATH: gno.land/r/crossrealm
package crossrealm

// A realm-embedding interface must carry realm's hidden seal through constant
// conversions. Checking only direct conversions to realm would leave this
// interface shape able to dispatch to attacker-controlled realm methods.
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

type R interface{ realm }

func main(cur realm) {
	r := R(fake(1))
	println(r.IsCurrent())
}

// Error:
// gno.land/r/crossrealm/zrealm_seal_constconv_iface.gno:25:7-17: gno.land/r/crossrealm.fake does not implement gno.land/r/crossrealm.R (missing method .uverse..seal)

// TypeCheckError:
// gno.land/r/crossrealm/zrealm_seal_constconv_iface.gno:25:9: cannot convert fake(1) (constant 1 of int type fake) to type R: fake does not implement R (missing method seal)
