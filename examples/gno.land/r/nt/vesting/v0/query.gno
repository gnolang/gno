package vesting

import (
	"chain/banker"

	vst "gno.land/p/nt/vesting/v0"

	"gno.land/r/nt/grc20reg/v0"
)

// Count is the number of grants created.
func Count() int { return len(grants) }

// GrantAddress is the sub-realm account that holds grant id.
func GrantAddress(id int64) address { return mustGrant(id).Address() }

// AdminOf is the grant's admin.
func AdminOf(id int64) address { return mustGrant(id).Admin() }

// BeneficiaryOf is the grant's beneficiary.
func BeneficiaryOf(id int64) address { return mustGrant(id).Beneficiary() }

// LabelOf is the grant's display name.
func LabelOf(id int64) string { return mustGrant(id).Label() }

// ClaimedOf is how much beneficiaries have taken.
func ClaimedOf(id int64) int64 { return mustGrant(id).Claimed() }

// ReturnedOf is how much a revoke sent back to the admin.
func ReturnedOf(id int64) int64 { return mustGrant(id).Returned() }

// RevokedOf reports whether the grant has been revoked.
func RevokedOf(id int64) bool { return mustGrant(id).Revoked() }

// RevocableOf reports whether the admin may revoke.
func RevocableOf(id int64) bool { return mustGrant(id).Revocable() }

// VestedOf is the amount unlocked at the current block time.
func VestedOf(id int64) int64 {
	n, err := mustGrant(id).Vested(nowUnix())
	must(err)
	return n
}

// ClaimableOf is the amount the beneficiary could claim right now.
func ClaimableOf(id int64) int64 {
	n, err := mustGrant(id).Releasable(nowUnix())
	must(err)
	return n
}

// OutstandingOf is what the books still expect on the grant address.
func OutstandingOf(id int64) int64 {
	n, err := mustGrant(id).Outstanding()
	must(err)
	return n
}

// BalanceOf reads the grant address. Coins use the bank. GRC20 uses the
// registry. A missing token reads as zero.
func BalanceOf(id int64) int64 {
	g := mustGrant(id)
	switch g.Asset().Kind {
	case vst.AssetCoin:
		return banker.NewReadonlyBanker().GetCoin(g.Address(), g.Asset().Denom)
	case vst.AssetGRC20:
		tok := grc20reg.Get(g.Asset().Token)
		if tok == nil {
			return 0
		}
		return tok.BalanceOf(g.Address())
	default:
		return 0
	}
}
