package grc721

import (
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
)

func newDummyRoyaltyNFT(_ int, rlm realm) *royaltyNFT {
	var nft *royaltyNFT
	func(cur realm) {
		nft = NewNFTWithRoyalty(0, cur, dummyNFTName, dummyNFTSymbol)
	}(cross(rlm))
	return nft
}

func TestSetTokenRoyalty(cur realm, t *testing.T) {
	dummy := newDummyRoyaltyNFT(0, cur)
	uassert.True(t, dummy != nil, "should not be nil")

	addr1 := testutils.TestAddress("alice")
	addr2 := testutils.TestAddress("bob")

	paymentAddress := testutils.TestAddress("john")
	percentage := int64(10) // 10%

	salePrice := int64(1000)
	expectRoyaltyAmount := int64(100)

	dummy.mint(addr1, TokenID("1"))

	derr := dummy.SetTokenRoyalty(addr1, TokenID("1"), RoyaltyInfo{
		PaymentAddress: paymentAddress,
		Percentage:     percentage,
	})
	uassert.NoError(t, derr, "Should not result in error")

	// Test case: Invalid token ID
	err := dummy.SetTokenRoyalty(addr1, TokenID("3"), RoyaltyInfo{
		PaymentAddress: paymentAddress,
		Percentage:     percentage,
	})
	uassert.ErrorIs(t, err, ErrInvalidTokenId)

	// Test case: Invalid payment address
	aerr := dummy.SetTokenRoyalty(addr1, TokenID("4"), RoyaltyInfo{
		PaymentAddress: address("###"), // invalid address
		Percentage:     percentage,
	})
	uassert.ErrorIs(t, aerr, ErrInvalidRoyaltyPaymentAddress)

	// Test case: Invalid percentage
	perr := dummy.SetTokenRoyalty(addr1, TokenID("5"), RoyaltyInfo{
		PaymentAddress: paymentAddress,
		Percentage:     int64(200), // over maxRoyaltyPercentage
	})
	uassert.ErrorIs(t, perr, ErrInvalidRoyaltyPercentage)

	// addr2 is not the owner.
	cerr := dummy.SetTokenRoyalty(addr2, TokenID("1"), RoyaltyInfo{
		PaymentAddress: paymentAddress,
		Percentage:     percentage,
	})
	uassert.ErrorIs(t, cerr, ErrCallerIsNotOwner)

	dummyPaymentAddress, dummyRoyaltyAmount, rerr := dummy.RoyaltyInfo(TokenID("1"), salePrice)
	uassert.NoError(t, rerr, "RoyaltyInfo error")
	uassert.Equal(t, paymentAddress, dummyPaymentAddress)
	uassert.Equal(t, expectRoyaltyAmount, dummyRoyaltyAmount)
}
