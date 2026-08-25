package banker

import (
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
)

// TestDepositRejectsNonUserCall pins the guard this page teaches: Deposit
// requires IsUserCall, so a realm-mediated call is refused even though it can
// carry an origin-send envelope.
func TestDepositRejectsNonUserCall(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/docs/zzcaller"))
	uassert.AbortsWithMessage(t, cur, "invalid non-user call", func(cur realm) {
		Deposit(cross(cur))
	})
}

// TestUserRunIsNotAUserCall is the reason the guard reads IsUserCall rather
// than IsUser: a `gnokey maketx run` script executes in an ephemeral
// /e/{addr}/run realm, which IsUser admits and IsUserCall refuses.
func TestUserRunIsNotAUserCall(t *testing.T) {
	addr := testutils.TestAddress("alice")
	run := testing.MakeRealm(addr, "gno.land/e/"+addr.String()+"/run", testing.OriginRealm())

	uassert.True(t, run.IsUserRun(), "the /e/{addr}/run shape is a user run")
	uassert.True(t, run.IsUser(), "IsUser admits it")
	uassert.False(t, run.IsUserCall(), "IsUserCall refuses it")
}
