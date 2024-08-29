package auction

import (
	"std"
	"testing"
	"time"
)

func setCurrentTime(t time.Time) {
	currentTime = t
}

func resetCurrentTime() {
	currentTime = time.Time{}
}

func TestNewAuction(t *testing.T) {
	owner := std.Address("owner")
	begin := time.Now().Add(1 * time.Hour)
	end := time.Now().Add(24 * time.Hour)
	startingBid := std.NewCoin("ugnot", 100)

	auction := NewAuction("Test Auction", owner, "This is a test auction", begin, end, startingBid)

	if auction.Title != "Test Auction" {
		t.Fatalf("expected auction title to be 'Test Auction', got '%s'", auction.Title)
	}
	if auction.Owner() != owner { // Note the change here to access owner
		t.Fatalf("expected auction owner to be '%s', got '%s'", owner, auction.Owner())
	}
	if auction.Description != "This is a test auction" {
		t.Fatalf("expected auction description to be 'This is a test auction', got '%s'", auction.Description)
	}
	if auction.Begin != begin {
		t.Fatalf("expected auction begin time to be '%s', got '%s'", begin, auction.Begin)
	}
	if auction.End != end {
		t.Fatalf("expected auction end time to be '%s', got '%s'", end, auction.End)
	}
	if !auction.StartingBid.IsEqual(startingBid) {
		t.Fatalf("expected auction starting bid to be '%s', got '%s'", startingBid.String(), auction.StartingBid.String())
	}
	if auction.State != "upcoming" {
		t.Fatalf("expected auction state to be 'upcoming', got '%s'", auction.State)
	}
}

func TestAddBid(t *testing.T) {
	owner := std.Address("owner")
	bidder1 := std.Address("bidder1")
	bidder2 := std.Address("bidder2")
	begin := time.Now().Add(1 * time.Hour)
	end := time.Now().Add(24 * time.Hour)
	startingBid := std.NewCoin("ugnot", 100)

	auction := NewAuction("Test Auction", owner, "This is a test auction", begin, end, startingBid)

	// Test before auction starts
	setCurrentTime(time.Now())
	err := auction.AddBid(bidder1, std.NewCoin("ugnot", 200))
	if err == nil || err.Error() != "auction: AddBid: auction has not started yet" {
		t.Fatalf("expected error 'auction has not started yet', got '%v'", err)
	}
	resetCurrentTime()

	// Test successful bid
	setCurrentTime(begin.Add(1 * time.Second))
	err = auction.AddBid(bidder1, std.NewCoin("ugnot", 200))
	resetCurrentTime()

	if err != nil {
		t.Fatalf("expected no error, got '%v'", err)
	}
	if !auction.StartingBid.IsEqual(std.NewCoin("ugnot", 200)) {
		t.Fatalf("expected auction starting bid to be '200ugnot', got '%s'", auction.StartingBid.String())
	}
	if len(auction.Bids) != 1 {
		t.Fatalf("expected number of bids to be '1', got '%d'", len(auction.Bids))
	}
	if auction.Bids[0].Bidder != bidder1 {
		t.Fatalf("expected bidder to be 'bidder1', got '%s'", auction.Bids[0].Bidder)
	}

	// Test higher bid
	setCurrentTime(begin.Add(2 * time.Second))
	err = auction.AddBid(bidder2, std.NewCoin("ugnot", 300))
	resetCurrentTime()

	if err != nil {
		t.Fatalf("expected no error, got '%v'", err)
	}
	if !auction.StartingBid.IsEqual(std.NewCoin("ugnot", 300)) {
		t.Fatalf("expected auction starting bid to be '300ugnot', got '%s'", auction.StartingBid.String())
	}
	if len(auction.Bids) != 2 {
		t.Fatalf("expected number of bids to be '2', got '%d'", len(auction.Bids))
	}
	if auction.Bids[1].Bidder != bidder2 {
		t.Fatalf("expected bidder to be 'bidder2', got '%s'", auction.Bids[1].Bidder)
	}

	// Test bid lower than current price
	setCurrentTime(begin.Add(3 * time.Second))
	err = auction.AddBid(bidder1, std.NewCoin("ugnot", 250))
	resetCurrentTime()
	if err == nil || err.Error() != "auction: AddBid: bid amount must be higher than the current highest bid" {
		t.Fatalf("expected error 'bid amount must be higher than the current highest bid', got '%v'", err)
	}
}

func TestEndAuction(t *testing.T) {
	owner := std.Address("owner")
	bidder := std.Address("bidder")
	begin := time.Now().Add(1 * time.Hour)
	end := time.Now().Add(24 * time.Hour)
	startingBid := std.NewCoin("ugnot", 100)

	auction := NewAuction("Test Auction", owner, "This is a test auction", begin, end, startingBid)

	// Test ending auction before it starts
	setCurrentTime(begin.Add(-1 * time.Hour))
	err := auction.EndAuction(owner)
	if err == nil || err.Error() != "auction: EndAuction: auction cannot end before the end time" {
		t.Fatalf("expected error 'auction cannot end before the end time', got '%v'", err)
	}
	resetCurrentTime()

	// Test ending auction with no bids
	setCurrentTime(end.Add(1 * time.Second))
	err = auction.EndAuction(owner)
	if err == nil || err.Error() != "auction: EndAuction: auction ended with no bids" {
		t.Fatalf("expected error 'auction ended with no bids', got '%v'", err)
	}
	resetCurrentTime()

	// Place a bid and end auction
	setCurrentTime(begin.Add(1 * time.Second))
	auction.AddBid(bidder, std.NewCoin("ugnot", 200))
	setCurrentTime(end.Add(1 * time.Second))
	err = auction.EndAuction(owner)
	resetCurrentTime()

	if err != nil {
		t.Fatalf("expected no error, got '%v'", err)
	}
	if auction.State != "closed" {
		t.Fatalf("expected auction state to be 'closed', got '%s'", auction.State)
	}
	if len(auction.Bids) == 0 {
		t.Fatalf("expected at least one bid to be present")
	}
	if auction.Bids[len(auction.Bids)-1].Bidder != bidder {
		t.Fatalf("expected winner to be 'bidder', got '%s'", auction.Bids[len(auction.Bids)-1].Bidder)
	}
}
