package grc721

import (
	"testing"

	"gno.land/p/nt/testutils"
	"gno.land/p/nt/uassert"
)

func TestSetTokenRoyalty(t *testing.T) {
	dummy := NewNFTWithRoyalty(dummyNFTName, dummyNFTSymbol)
	uassert.True(t, dummy != nil, "should not be nil")

	addr1 := testutils.TestAddress("alice")
	addr2 := testutils.TestAddress("bob")

	paymentAddress := testutils.TestAddress("john")
	percentage := int64(10) // 10%

	salePrice := int64(1000)
	expectRoyaltyAmount := int64(100)

	testing.SetOriginCaller(addr1) // addr1

	dummy.mint(addr1, TokenID("1"))

	derr := dummy.SetTokenRoyalty(TokenID("1"), RoyaltyInfo{
		PaymentAddress: paymentAddress,
		Percentage:     percentage,
	})
	uassert.NoError(t, derr, "Should not result in error")

	// Test case: Invalid token ID
	_ = dummy.SetTokenRoyalty(TokenID("3"), RoyaltyInfo{
		PaymentAddress: paymentAddress,
		Percentage:     percentage,
	})
	uassert.ErrorIs(t, derr, ErrInvalidTokenId)

	testing.SetOriginCaller(addr2) // addr2

	cerr := dummy.SetTokenRoyalty(TokenID("1"), RoyaltyInfo{
		PaymentAddress: paymentAddress,
		Percentage:     percentage,
	})
	uassert.ErrorIs(t, cerr, ErrCallerIsNotOwner)

	// Test case: Invalid payment address
	aerr := dummy.SetTokenRoyalty(TokenID("4"), RoyaltyInfo{
		PaymentAddress: address("###"), // invalid address
		Percentage:     percentage,
	})
	uassert.ErrorIs(t, aerr, ErrInvalidRoyaltyPaymentAddress)

	// Test case: Invalid percentage
	perr := dummy.SetTokenRoyalty(TokenID("5"), RoyaltyInfo{
		PaymentAddress: paymentAddress,
		Percentage:     int64(200), // over maxRoyaltyPercentage
	})
	uassert.ErrorIs(t, perr, ErrInvalidRoyaltyPercentage)

	// Test case: Retrieving Royalty Info
	testing.SetOriginCaller(addr1) // addr1

	dummyPaymentAddress, dummyRoyaltyAmount, rerr := dummy.RoyaltyInfo(TokenID("1"), salePrice)
	uassert.NoError(t, rerr, "RoyaltyInfo error")
	uassert.Equal(t, paymentAddress, dummyPaymentAddress)
	uassert.Equal(t, expectRoyaltyAmount, dummyRoyaltyAmount)
}
