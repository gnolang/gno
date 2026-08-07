package foo721

import (
	"testing"

	"gno.land/p/demo/tokens/grc721"
	"gno.land/p/nt/uassert/v0"
)

func TestFoo721(t *testing.T) {
	adminAddr := address("g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh")

	// init mints 10 tokens to the admin.
	uassert.Equal(t, int64(10), BalanceOf(adminAddr))
	uassert.Equal(t, adminAddr, OwnerOf(grc721.TokenID("0")))
	uassert.False(t, IsApprovedForAll(adminAddr, adminAddr))
}

func TestFoo721DefaultRoyalty(t *testing.T) {
	adminAddr := address("g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh")

	// The collection-level default is 5% to the admin, applied to every token.
	recv, amount := RoyaltyInfo(grc721.TokenID("0"), 10000)
	uassert.Equal(t, adminAddr, recv)
	uassert.Equal(t, int64(500), amount)
}
