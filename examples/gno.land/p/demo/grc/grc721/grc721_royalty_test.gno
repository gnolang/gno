package grc721

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/ufmt"
	"gno.land/p/demo/users"
)

func TestSetTokenRoyalty(t *testing.T) {
	dummy := NewNFTWithRoyalty(dummyNFTName, dummyNFTSymbol)
	if dummy == nil {
		t.Errorf("should not be nil")
	}

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

	if derr != nil {
		t.Errorf("Should not result in error : %s", derr.Error())
	}

	// Test case: Invalid token ID
	err := dummy.SetTokenRoyalty(TokenID("3"), RoyaltyInfo{
		PaymentAddress: paymentAddress,
		Percentage:     percentage,
	})
	if err != ErrInvalidTokenId {
		t.Errorf("Expected error %s, got %s", ErrInvalidTokenId, err)
	}

	std.TestSetOrigCaller(addr2) // addr2

	cerr := dummy.SetTokenRoyalty(TokenID("1"), RoyaltyInfo{
		PaymentAddress: paymentAddress,
		Percentage:     percentage,
	})
	if cerr != ErrCallerIsNotOwner {
		t.Errorf("Expected error %s, got %s", ErrCallerIsNotOwner, cerr)
	}

	// Test case: Invalid payment address
	aerr := dummy.SetTokenRoyalty(TokenID("4"), RoyaltyInfo{
		PaymentAddress: std.Address("###"), // invalid address
		Percentage:     percentage,
	})
	if aerr != ErrInvalidRoyaltyPaymentAddress {
		t.Errorf("Expected error %s, got %s", ErrInvalidRoyaltyPaymentAddress, aerr)
	}

	// Test case: Invalid percentage
	perr := dummy.SetTokenRoyalty(TokenID("5"), RoyaltyInfo{
		PaymentAddress: paymentAddress,
		Percentage:     uint64(200), // over maxRoyaltyPercentage
	})

	if perr != ErrInvalidRoyaltyPercentage {
		t.Errorf("Expected error %s, got %s", ErrInvalidRoyaltyPercentage, perr)
	}

	// Test case: Retrieving Royalty Info
	std.TestSetOrigCaller(addr1) // addr1

	dummyPaymentAddress, dummyRoyaltyAmount, rerr := dummy.RoyaltyInfo(TokenID("1"), salePrice)
	if rerr != nil {
		t.Errorf("RoyaltyInfo error: %s", rerr.Error())
	}

	if dummyPaymentAddress != paymentAddress {
		t.Errorf("Expected RoyaltyPaymentAddress %s, got %s", paymentAddress, dummyPaymentAddress)
	}

	if dummyRoyaltyAmount != expectRoyaltyAmount {
		t.Errorf("Expected RoyaltyAmount %d, got %d", expectRoyaltyAmount, dummyRoyaltyAmount)
	}
}
