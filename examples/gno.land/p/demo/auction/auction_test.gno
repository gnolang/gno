package auction

import (
	"std"
	"testing"
	"time"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
)

func TestNewAuction(t *testing.T) {
	owner := testutils.TestAddress("owner")
	begin := time.Now().Add(1 * time.Hour)
	end := time.Now().Add(24 * time.Hour)
	minPrice := std.NewCoin("ugnot", 100)
	auction := NewAuction("Test Auction", owner, "This is a test auction", begin, end, minPrice)

	uassert.Equal(t, "Test Auction", auction.Title, "Auction title mismatch")
	uassert.Equal(t, owner, auction.Owner.Owner(), "Auction owner mismatch")
	uassert.Equal(t, "This is a test auction", auction.Description, "Auction description mismatch")
	uassert.Equal(t, begin.Unix(), auction.Begin.Unix(), "Auction begin time mismatch")
	uassert.Equal(t, end.Unix(), auction.End.Unix(), "Auction end time mismatch")
	uassert.Equal(t, minPrice.Amount, auction.Price.Amount, "Auction amount price mismatch")
	uassert.Equal(t, minPrice.Denom, auction.Price.Denom, "Auction denom price mismatch")
}
func TestAuctionIsActive(t *testing.T) {
	owner := testutils.TestAddress("owner")
	begin := time.Now().Add(1 * time.Hour)
	end := time.Now().Add(24 * time.Hour)
	minPrice := std.NewCoin("ugnot", 100)
	auction := NewAuction("Test Auction", owner, "This is a test auction", begin, end, minPrice)

	// Test before auction starts
	err := auction.auctionIsActive()
	uassert.ErrorContains(t, err, "auction has not started yet", "Expected error for auction not started")

	// Simulate auction is active
	auction.Begin = time.Now().Add(-1 * time.Hour)
	auction.End = time.Now().Add(1 * time.Hour)
	err = auction.auctionIsActive()
	uassert.NoError(t, err, "Expected no error for auction active")

	// Simulate auction has ended
	auction.End = time.Now().Add(-1 * time.Hour)
	err = auction.auctionIsActive()
	uassert.ErrorContains(t, err, "auction has already ended", "Expected error for auction ended")
}
func TestValidateBidAmount(t *testing.T) {
	owner := testutils.TestAddress("owner")
	begin := time.Now().Add(1 * time.Hour)
	end := time.Now().Add(24 * time.Hour)
	minPrice := std.NewCoin("ugnot", 100)
	auction := NewAuction("Test Auction", owner, "This is a test auction", begin, end, minPrice)
	bidAmount := std.NewCoin("ugnot", 100)
	// Test valid bid amount
	err := auction.validateBidAmount(bidAmount)
	uassert.NoError(t, err, "Expected no error for valid bid amount")
	bidAmount = std.NewCoin("ugnot", 0)
	// Test invalid bid amount
	err = auction.validateBidAmount(bidAmount)
	uassert.ErrorContains(t, err, "bid amount must be positive", "Expected error for invalid bid amount")
}
func TestPlaceBid(t *testing.T) {
	owner := testutils.TestAddress("owner")
	bidder := testutils.TestAddress("bidder")
	begin := time.Now().Add(-1 * time.Hour) // Auction started 1 hour ago
	end := time.Now().Add(23 * time.Hour)   // Auction ends in 23 hours
	minPrice := std.NewCoin("ugnot", 100)
	auction := NewAuction("Test Auction", owner, "This is a test auction", begin, end, minPrice)

	// Place a bid
	bidAmount := std.NewCoin("ugnot", 150)
	auction.placeBid(bidder, bidAmount)

	// Verify the bid was added
	uassert.Equal(t, 1, len(auction.Bids), "Expected one bid in the auction")
	uassert.Equal(t, bidder, auction.Bids[0].Bidder, "Bidder address mismatch")
	uassert.Equal(t, bidAmount.Amount, auction.Bids[0].Amount.Amount, "Bid amount mismatch")
	uassert.Equal(t, bidAmount.Denom, auction.Bids[0].Amount.Denom, "Bid denom mismatch")

	// Verify the auction price was updated
	uassert.Equal(t, bidAmount.Amount, auction.Price.Amount, "Auction price mismatch")
	uassert.Equal(t, bidAmount.Denom, auction.Price.Denom, "Auction denom mismatch")
}

func TestAddBid(t *testing.T) {
	owner := testutils.TestAddress("owner")
	bidder1 := testutils.TestAddress("bidder1")
	bidder2 := testutils.TestAddress("bidder2")
	begin := time.Now().Add(-1 * time.Hour) // Auction started 1 hour ago
	end := time.Now().Add(23 * time.Hour)   // Auction ends in 23 hours
	minPrice := std.NewCoin("ugnot", 100)
	auction := NewAuction("Test Auction", owner, "This is a test auction", begin, end, minPrice)

	// Test placing a valid bid
	bidAmount1 := std.NewCoin("ugnot", 200)
	err := auction.AddBid(bidder1, bidAmount1)
	uassert.NoError(t, err, "Expected no error for valid bid")
	uassert.Equal(t, bidAmount1.Amount, auction.Price.Amount, "Auction price mismatch after first bid")
	uassert.Equal(t, 1, len(auction.Bids), "Expected one bid")
	uassert.Equal(t, bidder1, auction.Bids[0].Bidder, "Bidder mismatch")
	uassert.Equal(t, bidAmount1.Amount, auction.Bids[0].Amount.Amount, "Bid amount mismatch")

	// Test placing a higher bid
	bidAmount2 := std.NewCoin("ugnot", 300)
	err = auction.AddBid(bidder2, bidAmount2)
	uassert.NoError(t, err, "Expected no error for valid higher bid")
	uassert.Equal(t, bidAmount2.Amount, auction.Price.Amount, "Auction first price mismatch after second bid")
	uassert.Equal(t, 2, len(auction.Bids), "Expected two bids")
	uassert.Equal(t, bidder2, auction.Bids[1].Bidder, "Bidder mismatch")
	uassert.Equal(t, bidAmount2.Amount, auction.Bids[1].Amount.Amount, "Bid amount mismatch")

	// Test placing a bid lower than the current highest bid
	bidAmount3 := std.NewCoin("ugnot", 250)
	err = auction.AddBid(bidder1, bidAmount3)
	uassert.ErrorContains(t, err, "bid amount must be higher than the last bid", "Expected error for lower bid")

	// Test placing a bid when the auction is not active (before it starts)
	auction.Begin = time.Now().Add(1 * time.Hour) // Auction starts in 1 hour
	err = auction.AddBid(bidder1, bidAmount1)
	uassert.ErrorContains(t, err, "auction has not started yet", "Expected error for auction not started")

	// Test placing a bid when the auction is not active (after it ends)
	auction.Begin = time.Now().Add(-2 * time.Hour) // Auction started 2 hours ago
	auction.End = time.Now().Add(-1 * time.Hour)   // Auction ended 1 hour ago
	err = auction.AddBid(bidder1, bidAmount1)
	uassert.ErrorContains(t, err, "auction has already ended", "Expected error for auction ended")
}

func TestEndAuction(t *testing.T) {
	owner := testutils.TestAddress("owner")
	bidder := testutils.TestAddress("bidder")
	begin := time.Now().Add(-2 * time.Hour) // Set begin time to the past
	end := time.Now().Add(-1 * time.Hour)   // Set end time to the past
	minPrice := std.NewCoin("ugnot", 100)

	auction := NewAuction("Test Auction", owner, "This is a test auction", begin, end, minPrice)

	// Simulate the auction running and place a bid
	std.TestSetOrigCaller(bidder)
	bidAmount := std.NewCoin("ugnot", 150)
	auction.Begin = time.Now().Add(-2 * time.Hour) // Ensure auction has started
	auction.End = time.Now().Add(1 * time.Hour)    // Ensure auction is not ended
	err := auction.AddBid(bidder, bidAmount)
	uassert.NoError(t, err, "Unexpected error when placing a bid")

	// Set the auction end time to the past to simulate the auction ending
	auction.End = time.Now().Add(-1 * time.Second)

	// End the auction
	std.TestSetOrigCaller(owner)
	err = auction.EndAuction()

	// Ensure no error occurs when ending the auction with bids
	uassert.NoError(t, err, "Expected no error when auction ends with bids")

	// Check that the winner is the correct bidder
	if len(auction.Bids) > 0 {
		highestBid := auction.Price.Amount
		winner := auction.Bids[len(auction.Bids)-1].Bidder
		uassert.Equal(t, bidder, winner, "Winner should be the highest bidder")

		// Check that the highest bid is 150 tokens
		uassert.Equal(t, bidAmount.Amount, highestBid, "Highest bid should be 150 ")
	} else {
		t.Fatalf("Expected at least one bid, but got none")
	}
}