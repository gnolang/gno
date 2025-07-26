package auction

import (
	"std"
	"testing"
	"time"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
)

var (
	user1 = testutils.TestAddress("user1")
	user2 = testutils.TestAddress("user2")
)

func TestCreateAuction(t *testing.T) {
	// Initialize the test environment for user1
	std.TestSetRealm(std.NewUserRealm(user1))

	// Create a new auction
	title := "Test Auction"
	description := "This is a test auction"
	begin := time.Now().Unix()
	end := time.Now().Add(time.Hour * 24).Unix()
	amount := int64(100)
	denom := "ugnot"

	result := CreateAuction(title, description, begin, end, amount, denom)
	uassert.Equal(t, "Auction created successfully.", result)
	uassert.Equal(t, 1, len(auctionList)) // Check that the auction was added to the list

	// Check the details of the created auction
	auction := auctionList[0]
	uassert.Equal(t, title, auction.Title)
	uassert.Equal(t, description, auction.Description)
	uassert.Equal(t, begin, auction.Begin.Unix())
	uassert.Equal(t, end, auction.End.Unix())
	uassert.Equal(t, amount, auction.Price.Amount)
	uassert.Equal(t, denom, auction.Price.Denom)
}
func TestPlaceBid(t *testing.T) {
	// Initialize the environment for user1 as the auction creator
	std.TestSetOrigCaller(user1)
	CreateAuction("Test Auction", "This is a test auction", time.Now().Unix(), time.Now().Add(time.Hour*24).Unix(), 100, "ugnot")

	// Switch to user2 as the bidder
	std.TestSetOrigCaller(user2)                                       // Simulate user2 as the caller
	std.TestSetOrigSend(std.Coins{{Denom: "ugnot", Amount: 200}}, nil) // Set user2's funds for the transaction
	// Place a bid of 150 ugnots
	bidAmount := int64(150)
	denom := "ugnot"
	result := PlaceBid(0, bidAmount, denom)
	uassert.Equal(t, "Bid placed successfully", result)

	// Verify that the bid was successfully recorded in the auction
	auction := auctionList[0]
	lastBid := auction.Bids[len(auction.Bids)-1]
	uassert.Equal(t, bidAmount, lastBid.Amount.Amount)
	uassert.Equal(t, user2, lastBid.Bidder)
}

func TestEndAuction(t *testing.T) {
	// Set user1 as the auction creator
	std.TestSetOrigCaller(user1)
	CreateAuction("Test Auction", "This is a test auction", time.Now().Add(-2*time.Hour).Unix(), time.Now().Add(1*time.Hour).Unix(), 100, "ugnot")

	// Place a bid with user2 before auction ends
	std.TestSetOrigCaller(user2)
	std.TestIssueCoins(user2, std.Coins{{Denom: "ugnot", Amount: 200}})
	std.TestSetOrigSend(std.Coins{{Denom: "ugnot", Amount: 150}}, nil)

	// Place the bid while auction is active
	result := PlaceBid(0, 150, "ugnot")
	uassert.Equal(t, "Bid placed successfully", result)

	// Simulate time progression after the bid
	std.TestSkipHeights(2000) // Advance time to pass the auction's end time

	// End the auction
	std.TestSetOrigCaller(user1) // Set user1's realm
	result = EndAuction(0)
	uassert.Equal(t, "Auction ended successfully", result)

	// Verify the highest bidder
	auction := auctionList[0]
	highestBid := auction.Bids[len(auction.Bids)-1]
	uassert.Equal(t, user2, highestBid.Bidder)
	uassert.Equal(t, int64(150), highestBid.Amount.Amount)
}
