package grc721

import (
	"math"
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
)

func TestSetTokenRoyalty(t *testing.T) {
	dummy := NewNFTWithRoyalty(dummyNFTName, dummyNFTSymbol)
	uassert.True(t, dummy != nil, "should not be nil")

	addr1 := testutils.TestAddress("alice")
	addr2 := testutils.TestAddress("bob")

	paymentAddress := testutils.TestAddress("john")
	bps := int64(1000) // 10% in basis points

	salePrice := int64(1000)
	expectRoyaltyAmount := int64(100)

	testing.SetOriginCaller(addr1) // addr1

	dummy.mint(addr1, TokenID("1"))

	derr := dummy.SetTokenRoyalty(TokenID("1"), RoyaltyInfo{
		PaymentAddress: paymentAddress,
		Bps:            bps,
	})
	uassert.NoError(t, derr, "Should not result in error")

	// Test case: Invalid token ID
	derr = dummy.SetTokenRoyalty(TokenID("3"), RoyaltyInfo{
		PaymentAddress: paymentAddress,
		Bps:            bps,
	})
	uassert.ErrorIs(t, derr, ErrInvalidTokenId)

	testing.SetOriginCaller(addr2) // addr2

	cerr := dummy.SetTokenRoyalty(TokenID("1"), RoyaltyInfo{
		PaymentAddress: paymentAddress,
		Bps:            bps,
	})
	uassert.ErrorIs(t, cerr, ErrCallerIsNotOwner)

	// Test case: Invalid payment address
	aerr := dummy.SetTokenRoyalty(TokenID("4"), RoyaltyInfo{
		PaymentAddress: address("###"), // invalid address
		Bps:            bps,
	})
	uassert.ErrorIs(t, aerr, ErrInvalidRoyaltyPaymentAddress)

	// Test case: bps above maxRoyaltyBps
	perr := dummy.SetTokenRoyalty(TokenID("5"), RoyaltyInfo{
		PaymentAddress: paymentAddress,
		Bps:            int64(10001), // over maxRoyaltyBps (10000)
	})
	uassert.ErrorIs(t, perr, ErrInvalidRoyaltyBps)

	// Test case: negative bps
	nerr := dummy.SetTokenRoyalty(TokenID("1"), RoyaltyInfo{
		PaymentAddress: paymentAddress,
		Bps:            int64(-1),
	})
	uassert.ErrorIs(t, nerr, ErrInvalidRoyaltyBps)

	// Test case: Retrieving Royalty Info
	testing.SetOriginCaller(addr1) // addr1

	dummyPaymentAddress, dummyRoyaltyAmount, rerr := dummy.RoyaltyInfo(TokenID("1"), salePrice)
	uassert.NoError(t, rerr, "RoyaltyInfo error")
	uassert.Equal(t, paymentAddress, dummyPaymentAddress)
	uassert.Equal(t, expectRoyaltyAmount, dummyRoyaltyAmount)
}

func TestRoyaltyBpsCompliance(t *testing.T) {
	owner := testutils.TestAddress("owner")
	payee := testutils.TestAddress("payee")

	nft := NewNFTWithRoyalty(dummyNFTName, dummyNFTSymbol)
	testing.SetOriginCaller(owner)
	nft.mint(owner, TokenID("1"))

	tests := []struct {
		name       string
		bps        int64
		salePrice  int64
		wantAmount int64
	}{
		{"2.5% on 1000", 250, 1000, 25},
		{"0.5% on 1000", 50, 1000, 5},
		{"7.5% on 1000", 750, 1000, 75},
		{"10% on 1000", 1000, 1000, 100},
		{"100% on 1000", 10000, 1000, 1000},
		{"0% on 1000", 0, 1000, 0},
		// Truncation: 10% of 1 is 0.1, which rounds down to 0 in int64.
		{"10% on 1 truncates", 1000, 1, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := nft.SetTokenRoyalty(TokenID("1"), RoyaltyInfo{
				PaymentAddress: payee,
				Bps:            tt.bps,
			})
			uassert.NoError(t, err, "bps value must be accepted")

			_, got, _ := nft.RoyaltyInfo(TokenID("1"), tt.salePrice)
			uassert.Equal(t, got, tt.wantAmount)
		})
	}
}

// TestRoyaltyInfoOverflow ensures RoyaltyInfo does not panic when the
// precise product (salePrice * Bps) would overflow int64, and that the
// fallback path returns an approximate result close to the true royalty.
func TestRoyaltyInfoOverflow(t *testing.T) {
	owner := testutils.TestAddress("owner")
	payee := testutils.TestAddress("payee")

	nft := NewNFTWithRoyalty(dummyNFTName, dummyNFTSymbol)
	testing.SetOriginCaller(owner)
	nft.mint(owner, TokenID("1"))

	// 10% of MaxInt64 overflows when computed as MaxInt64 * 1000 first.
	err := nft.SetTokenRoyalty(TokenID("1"), RoyaltyInfo{
		PaymentAddress: payee,
		Bps:            int64(1000),
	})
	uassert.NoError(t, err)

	_, got, rerr := nft.RoyaltyInfo(TokenID("1"), int64(math.MaxInt64))
	uassert.NoError(t, rerr)
	// Fallback: (MaxInt64 / 10000) * 1000.
	want := (int64(math.MaxInt64) / bpsDenominator) * int64(1000)
	uassert.Equal(t, got, want)
}
