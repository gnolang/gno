package auction

import (
	"std"
	"strings"
	"testing"
	"time"

	auctionp "gno.land/p/demo/auction"
)

func TestAuction(t *testing.T) {
	// Simulate the admin creating an auction
	adminAddr := std.Address("admin")

	// Set the caller to admin
	std.TestSetOrigCaller(adminAddr)

	// Create an auction that starts almost immediately
	begin := time.Now().Unix()                   // Auction begins in 1 second
	end := time.Now().Add(24 * time.Hour).Unix() // Auction ends in 24 hours
	result := NewAuction("Test Auction", "A simple test auction", begin, end, 100)
	if result != "auction created" {
		t.Errorf("Auction should be created")
	}
	if auctionTree.Size() != 1 {
		t.Errorf("Auction should be created")
	}

	// Wait for the auction to start
	// time.Sleep(2 * time.Second)

	// Place bids by different users
	user1 := std.Address("user1")
	user2 := std.Address("user2")

	// Set the caller to user1 and place a bid
	std.TestSetOrigCaller(user1)
	result = AddBid("0", 200)
	if result != "bid placed" {
		t.Errorf("Bid should be placed")
	}

	// Set the caller to user2 and place a bid
	std.TestSetOrigCaller(user2)
	result = AddBid("0", 300)
	if result != "bid placed" {
		t.Errorf("Bid should be placed")
	}

	// Ensure the highest bid is correct
	auc, _ := auctionTree.Get("0")
	auction := auc.(*auctionp.Auction)
	if auction.GetPrice() != 300 {
		t.Errorf("Highest bid should be 300")
	}

	// Render the state
	state := Render("")
	if !strings.Contains(state, "Test Auction") {
		t.Errorf("Auction should be in the rendered state")
	}
}
