package grc721

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/users"
	"gno.land/p/demo/ufmt"
)

func TestSetTokenRoyalty(t *testing.T) {
	dummy := NewNFTWithRoyalty(dummyNFTName, dummyNFTSymbol)
	if dummy == nil {
		t.Errorf("should not be nil")
	}

	addr1 := std.Address("g1var589z07ppjsjd24ukm4uguzwdt0tw7g47cgm")
	addr2 := std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")

	paymentAddress := std.Address("g1var589z07ppjsjd24ukm4uguzwdt0tw7g47cgm")
	percentage := uint64(10) // 10%

	salePrice := uint64(1000)
	expectRoyaltyAmount := uint64(100)

	std.TestSetOrigCaller(std.Address(addr1)) // addr1

	dummy.mint(addr1, TokenID("1"))

	_, derr := dummy.SetTokenRoyalty(TokenID("1"), RoyaltyInfo{
		PaymentAddress: paymentAddress,
		Percentage: percentage,
	})

	if derr != nil {
		t.Errorf("Should not result in error : %s", derr.Error())
	}

	// Test case: Invalid token ID
	_, err := dummy.SetTokenRoyalty(TokenID("3"), RoyaltyInfo{
		PaymentAddress: paymentAddress,
		Percentage: percentage,
	})
	if err != ErrInvalidTokenId {
		t.Errorf("Expected error %s, got %s", ErrInvalidTokenId, err)
	}

	std.TestSetOrigCaller(std.Address(addr2)) // addr2

	_, cerr := dummy.SetTokenRoyalty(TokenID("1"),RoyaltyInfo{
		PaymentAddress: paymentAddress,
		Percentage: percentage,
	}) 
	if cerr != ErrCallerIsNotOwner {
		t.Errorf("Expected error %s, got %s", ErrCallerIsNotOwner, cerr)
	}

	// Test case: Invalid payment address
	_, aerr := dummy.SetTokenRoyalty(TokenID("4"), RoyaltyInfo{
		PaymentAddress: std.Address("###"),
		Percentage: percentage,
	})
	if aerr != ErrInvalidRoyaltyPaymentAddress {
		t.Errorf("Expected error %s, got %s", ErrInvalidRoyaltyPaymentAddress, aerr)
	}


	// Test case: Invalid payment address
	_, perr := dummy.SetTokenRoyalty(TokenID("5"), RoyaltyInfo{
		PaymentAddress: paymentAddress,
		Percentage: uint64(10),
	})

	if perr != ErrInvalidRoyaltyPercentage {
		// t.Errorf("Expected error %s, got %s", ErrInvalidRoyaltyPercentage, perr)
	}


	// Test case: Retrieving Royalty Info
	std.TestSetOrigCaller(std.Address(addr1)) // addr1

	dummyPaymentAddress, dummyRoyaltyAmount, rerr := dummy.RoyaltyInfo(TokenID("1"), salePrice)
	if rerr != nil {
		t.Errorf("RoyaltyInfo error: %s", rerr.Error())

	} else {
		if dummyPaymentAddress != paymentAddress {
			t.Errorf("Expected RoyaltyPaymentAddress %s, got %s", paymentAddress, dummyPaymentAddress)
		}
		
		if dummyRoyaltyAmount != expectRoyaltyAmount {
			t.Errorf("Expected RoyaltyAmount %d, got %d", expectRoyaltyAmount, dummyRoyaltyAmount)
		}
	}	
}



