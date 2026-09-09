package tba

import (
	"testing"

	"chain"

	"gno.land/p/demo/tokens/grc721"
	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
)

var (
	aliceAddr = testutils.TestAddress("alice")
	bobAddr   = testutils.TestAddress("bob")
)

// setup makes the realm usable from a unit test and mints one token to
// alice, returning its id. Uses cross(cur) for the crossing calls into
// Setup/Mint, matching how any external caller invokes them.
func setup(cur realm, t *testing.T) grc721.TokenID {
	t.Helper()
	testing.SetRealm(testing.NewUserRealm(admin))
	if nft == nil {
		Setup(cross(cur))
	}
	return Mint(cross(cur), aliceAddr, "art", "ipfs://img")
}

func TestSetupOnlyAdmin(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(bobAddr))
	uassert.AbortsContains(t, cur, "only admin can initialize", func(cur realm) {
		Setup(cross(cur))
	}, "non-admin Setup should abort")
}

func TestMintOnlyAdmin(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(admin))
	if nft == nil {
		Setup(cross(cur))
	}

	testing.SetRealm(testing.NewUserRealm(bobAddr))
	uassert.AbortsContains(t, cur, "only admin can mint", func(cur realm) {
		Mint(cross(cur), bobAddr, "art", "ipfs://img")
	}, "non-admin Mint should abort")
}

// TestTransferFromSucceedsForOwner: the owner of a token must be able to
// transfer it. This is the case that failed in the earlier draft test
// (address mismatch in the test's own expectations) — re-verified here
// against the actual owner returned by OwnerOf, not a hardcoded literal.
func TestTransferFromSucceedsForOwner(cur realm, t *testing.T) {
	tid := setup(cur, t)

	ownerBefore, err := OwnerOf(tid)
	uassert.NoError(t, err, "OwnerOf should not error")
	uassert.Equal(t, aliceAddr.String(), ownerBefore.String(), "alice should own the freshly minted token")

	testing.SetRealm(testing.NewUserRealm(aliceAddr))
	TransferFrom(cross(cur), aliceAddr, bobAddr, tid)

	ownerAfter, err := OwnerOf(tid)
	uassert.NoError(t, err, "OwnerOf should not error after transfer")
	uassert.Equal(t, bobAddr.String(), ownerAfter.String(), "bob should own the token after owner-initiated transfer")
}

// TestTransferFromAbortsWhenUnauthorized: a caller who is neither the owner
// nor an approved operator must not be able to move the token. The realm's
// TransferFrom panics on the underlying error instead of returning it, so
// the transaction aborts rather than silently committing a no-op.
func TestTransferFromAbortsWhenUnauthorized(cur realm, t *testing.T) {
	tid := setup(cur, t)

	testing.SetRealm(testing.NewUserRealm(bobAddr))
	uassert.AbortsContains(t, cur, "caller is not token owner or approved", func(cur realm) {
		TransferFrom(cross(cur), aliceAddr, bobAddr, tid)
	}, "unauthorized transfer should abort")

	owner, err := OwnerOf(tid)
	uassert.NoError(t, err, "OwnerOf should not error")
	uassert.Equal(t, aliceAddr.String(), owner.String(), "ownership must be unchanged after a rejected transfer")
}

func TestWithdrawRejectsNonPositive(cur realm, t *testing.T) {
	tid := setup(cur, t)
	testing.IssueCoins(VaultAddress(tid), chain.NewCoins(chain.NewCoin("ugnot", 1000)))

	testing.SetRealm(testing.NewUserRealm(aliceAddr))
	uassert.AbortsContains(t, cur, "amount must be positive", func(cur realm) {
		Withdraw(cross(cur), tid, bobAddr, 0)
	}, "zero-amount withdraw should abort")

	uassert.Equal(t, int64(1000), Balance(tid), "vault balance must be unchanged after a rejected withdraw")
}

func TestWithdrawOnlyOwner(cur realm, t *testing.T) {
	tid := setup(cur, t)
	testing.IssueCoins(VaultAddress(tid), chain.NewCoins(chain.NewCoin("ugnot", 1000)))

	testing.SetRealm(testing.NewUserRealm(bobAddr))
	uassert.AbortsContains(t, cur, "caller is not the current owner", func(cur realm) {
		Withdraw(cross(cur), tid, bobAddr, 100)
	}, "non-owner withdraw should abort")

	uassert.Equal(t, int64(1000), Balance(tid), "vault balance must be unchanged after a rejected withdraw")
}

func TestWithdrawSucceedsForOwner(cur realm, t *testing.T) {
	tid := setup(cur, t)
	testing.IssueCoins(VaultAddress(tid), chain.NewCoins(chain.NewCoin("ugnot", 1000)))

	testing.SetRealm(testing.NewUserRealm(aliceAddr))
	Withdraw(cross(cur), tid, bobAddr, 400)

	uassert.Equal(t, int64(600), Balance(tid), "vault balance should decrease by the withdrawn amount")
}

// TestVaultAddressDeterministic: VaultAddress must derive from the realm's
// actual deploy path (captured in Setup), not a hardcoded constant — this
// is the fix for the "copy advertises the original realm's vault" issue.
// We can't deploy a second copy from a unit test, but we can at least
// confirm VaultAddress is deterministic and non-empty for the realm under
// test.
func TestVaultAddressDeterministic(cur realm, t *testing.T) {
	tid := setup(cur, t)
	v1 := VaultAddress(tid)
	v2 := VaultAddress(tid)
	uassert.Equal(t, v1.String(), v2.String(), "VaultAddress must be deterministic for the same token")
	uassert.True(t, v1.String() != "", "VaultAddress must not be empty")
}
