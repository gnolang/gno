package grc721

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
)

func TestSetTokenRoyalty(t *testing.T) {
	dummy := NewNFTWithRoyalty(dummyNFTName, dummyNFTSymbol)
	uassert.True(t, dummy != nil, "should not be nil")

	addr1 := testutils.TestAddress("alice")
	addr2 := testutils.TestAddress("bob")

	paymentAddress := testutils.TestAddress("john")
	percentage := uint64(10) // 10%

	salePrice := uint64(1000)
	expectRoyaltyAmount := uint64(100)

	std.TestSetOrigCaller(addr1) // addr1

	dummy.mint(addr1, TokenID("1"))

	derr := dummy.SetTokenRoyalty(TokenID("1"), RoyaltyInfo{
		PaymentAddress: paymentAddress,
		Percentage:     percentage,
	})
	uassert.NoError(t, derr, "Should not result in error")

	// Test case: Invalid token ID
	err := dummy.SetTokenRoyalty(TokenID("3"), RoyaltyInfo{
		PaymentAddress: paymentAddress,
		Percentage:     percentage,
	})
	uassert.ErrorIs(t, derr, ErrInvalidTokenId)

	std.TestSetOrigCaller(addr2) // addr2

	cerr := dummy.SetTokenRoyalty(TokenID("1"), RoyaltyInfo{
		PaymentAddress: paymentAddress,
		Percentage:     percentage,
	})
	uassert.ErrorIs(t, cerr, ErrCallerIsNotOwner)

	// Test case: Invalid payment address
	aerr := dummy.SetTokenRoyalty(TokenID("4"), RoyaltyInfo{
		PaymentAddress: std.Address("###"), // invalid address
		Percentage:     percentage,
	})
	uassert.ErrorIs(t, aerr, ErrInvalidRoyaltyPaymentAddress)

	// Test case: Invalid percentage
	perr := dummy.SetTokenRoyalty(TokenID("5"), RoyaltyInfo{
		PaymentAddress: paymentAddress,
		Percentage:     uint64(200), // over maxRoyaltyPercentage
	})
	uassert.ErrorIs(t, perr, ErrInvalidRoyaltyPercentage)

	// Test case: Retrieving Royalty Info
	std.TestSetOrigCaller(addr1) // addr1

	dummyPaymentAddress, dummyRoyaltyAmount, rerr := dummy.RoyaltyInfo(TokenID("1"), salePrice)
	uassert.NoError(t, rerr, "RoyaltyInfo error")
	uassert.Equal(t, paymentAddress, dummyPaymentAddress)
	uassert.Equal(t, expectRoyaltyAmount, dummyRoyaltyAmount)
}
