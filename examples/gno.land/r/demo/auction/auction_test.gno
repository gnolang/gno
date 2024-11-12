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
	// Set user1 as the original caller and create the auction with both start and end times in the past
	std.TestSetRealm(std.NewUserRealm(user1))
	startTime := time.Now().Add(-2 * time.Hour) // Start time is 2 hours ago
	endTime := time.Now().Add(1 * time.Hour)    // End time is 1 hour ago
	CreateAuction("Test Auction", "This is a test auction", startTime.Unix(), endTime.Unix(), 100, "ugnot")

	// Switch context to user2 to place a bid
	std.TestSetOrigCaller(user2)
	std.TestSetOrigSend(std.Coins{{Denom: "ugnot", Amount: 200}}, nil)
	PlaceBid(0, 150, "ugnot")

	// Reset the context to user1 before ending the auction
	std.TestSetRealm(std.NewUserRealm(user1))
	auction := auctionList[0]
	auction.End = time.Now().Add(-1 * time.Hour) // Set the end time to 1 hour ago
	// End the auction and check for successful completion
	result := EndAuction(0)
	uassert.Equal(t, "Auction ended successfully", result)

	// Confirm that user2 is recorded as the highest bidder

	uassert.Equal(t, user2, auction.Bids[len(auction.Bids)-1].Bidder)
}
