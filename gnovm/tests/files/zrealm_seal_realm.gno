// PKGPATH: gno.land/r/crossrealm
package crossrealm

// User type that structurally implements every public method of the
// uverse `realm` interface (Address, PkgPath, Previous, IsCode, IsUser,
// IsUserCall, IsUserRun, IsEphemeral, IsCurrent, String). Before the
// seal, this would have been assignable to a `realm` variable and could
// have been passed wherever a captured `cur` was expected. The hidden
// `.seal` method (dot-named — undeclarable in source) is what makes
// satisfaction impossible from user code.

type fakeRealm struct{}

func (fakeRealm) Address() address  { return address("") }
func (fakeRealm) PkgPath() string   { return "" }
func (fakeRealm) Previous() realm   { return nil }
func (fakeRealm) IsCode() bool      { return false }
func (fakeRealm) IsUser() bool      { return false }
func (fakeRealm) IsUserCall() bool  { return false }
func (fakeRealm) IsUserRun() bool   { return false }
func (fakeRealm) IsEphemeral() bool { return false }
func (fakeRealm) IsCurrent() bool   { return false }
func (fakeRealm) String() string    { return "fake" }

var _ realm = fakeRealm{}

func main(cur realm) {}

// Error:
// gno.land/r/crossrealm/zrealm_seal_realm.gno:25:5-26: gno.land/r/crossrealm.fakeRealm does not implement .uverse.realm (missing method .seal)

// TypeCheckError:
// gno.land/r/crossrealm/zrealm_seal_realm.gno:25:15: cannot use fakeRealm{} (value of struct type fakeRealm) as realm value in variable declaration: fakeRealm does not implement gno0p9.realm (missing method Sub)
