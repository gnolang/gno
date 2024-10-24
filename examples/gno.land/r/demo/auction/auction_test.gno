package auction

import (
	"std"
	"strings"
	"testing"
	"time"
)

func setCurrentTime(t time.Time) {
	currentTime = t
}

func resetCurrentTime() {
	currentTime = time.Time{}
}

func TestAuction(t *testing.T) {
	// Initialize the router and handlers

	// Simulate the admin creating an auction
	adminAddr := std.Address("admin")
	std.TestSetOrigCaller(adminAddr)

	// Create an auction that starts almost immediately
	begin := time.Now().Add(1 * time.Second).Unix() // Auction begins in 1 second
	end := time.Now().Add(24 * time.Hour).Unix()    // Auction ends in 24 hours
	CreateAuction("Test Auction", "A simple test auction", begin, end, 100)

	// Check if auction is in the upcoming section
	std.TestSkipHeights(1) // Skip 1 block to simulate time passage
	updateAuctionStates()
	upcomingPage := Render("upcoming")
	if !strings.Contains(upcomingPage, "Test Auction") {
		t.Errorf("Auction should be listed in upcoming auctions")
	}

	// Simulate time passing to start the auction
	std.TestSkipHeights(360) // Skip 360 blocks (1800 seconds or 30 minutes)
	updateAuctionStates()

	// Check if auction is in the ongoing section
	ongoingPage := Render("ongoing")
	if !strings.Contains(ongoingPage, "Test Auction") {
		t.Errorf("Auction should be listed in ongoing auctions")
	}

	// Simulate users placing bids
	user1 := std.Address("user1")
	user2 := std.Address("user2")

	// Set the caller to user1 and place a bid
	std.TestSetOrigCaller(user1)
	PlaceBid(0, 200)

	// Set the caller to user2 and place a bid
	std.TestSetOrigCaller(user2)
	PlaceBid(0, 300)

	// Check the details of the auction to verify bids
	auctionDetails := Render("auction/0")
	if !strings.Contains(auctionDetails, "300") {
		t.Errorf("Highest bid should be 300")
	}

	// End the auction
	std.TestSetOrigCaller(adminAddr)
	EndAuction(0)

	// Check if auction is in the closed state
	std.TestSkipHeights(8640) // Skip 8640 blocks (43200 seconds or 12 hours)
	updateAuctionStates()
	auctionDetails = Render("auction/0")
	if !strings.Contains(auctionDetails, "Auction ended") {
		t.Errorf("Auction should be ended")
	}

	resetCurrentTime()
}

// Update the auction states based on the current time

func updateAuctionStates() {
	now := time.Now()
	for _, auc := range auctionList {
		if auc.State == "upcoming" && now.After(auc.Begin) {
			auc.State = "ongoing"
		} else if auc.State == "ongoing" && now.After(auc.End) {
			auc.State = "closed"
		}
	}
}
