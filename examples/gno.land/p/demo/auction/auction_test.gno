package auction

import (
	"std"
	"testing"
	"time"

	"gno.land/p/demo/testutils"
)

// TestBid_new verifies that a new Bid can be initialized correctly
func TestBid_new(t *testing.T) {
	var (
		amount = uint64(100)
		bidder = testutils.TestAddress("bidder")
	)

	// Create a new Bid instance
	B := NewBid(bidder, amount)
	if B.GetAmount() != amount {
		t.Fatalf("invalid amount")
	}
	if B.GetBidder() != bidder {
		t.Fatalf("invalid bidder")
	}
}

// TestAuction_new verifies that a new Auction can be initialized correctly
func TestAuction_new(t *testing.T) {
	var (
		title       = "auction title"
		owner       = testutils.TestAddress("owner")
		description = "description"
		begin       = time.Now()
		end         = begin.Add(time.Hour * 24)
		minPrice    = uint64(1)
	)

	// Create a new Auction instance
	A := NewAuction(title, owner, description, begin, end, minPrice)
	if A.GetTitle() != title {
		t.Fatalf("invalid title")
	}
	if A.GetOwner() != owner {
		t.Fatalf("invalid owner")
	}
	if A.GetDescription() != description {
		t.Fatalf("invalid description")
	}
	if A.GetBegin() != begin {
		t.Fatalf("invalid begin")
	}
	if A.GetEnd() != end {
		t.Fatalf("invalid end")
	}
	if A.GetPrice() != minPrice {
		t.Fatalf("invalid minPrice")
	}

	std.TestSetOrigCaller(owner)
	if !A.IsOwner(owner) {
		t.Fatalf("invalid owner check")
	}
}

// TestAuction_AddBid verifies that bids can be added to an auction correctly
func TestAuction_AddBid(t *testing.T) {
	var (
		title       = "auction title"
		owner       = testutils.TestAddress("owner")
		bidder1     = testutils.TestAddress("bidder1")
		bidder2     = testutils.TestAddress("bidder2")
		description = "description"
		begin       = time.Now()
		end         = begin.Add(time.Hour * 24)
		minPrice    = uint64(100)
	)

	// Create a new Auction instance
	A := NewAuction(title, owner, description, begin, end, minPrice)

	// to do Simulate time to be after auction start
	//  std.TestSetTime(begin.Add(time.Second))

	// Place first bid
	std.TestSetOrigCaller(bidder1)
	A.AddBid(bidder1, 200)
	if A.GetPrice() != 200 {
		t.Fatalf("First bid failed")
	}

	// Place second bid
	std.TestSetOrigCaller(bidder2)
	A.AddBid(bidder2, 300)
	if A.GetPrice() != 300 {
		t.Fatalf("Second bid failed")
	}

	// Verify bids
	bids := A.GetBids()
	if len(bids) != 2 {
		t.Fatalf("Expected 2 bids, got %d", len(bids))
	}
	if bids[0].GetBidder() != bidder1 || bids[0].GetAmount() != 200 {
		t.Fatalf("First bid incorrect")
	}
	if bids[1].GetBidder() != bidder2 || bids[1].GetAmount() != 300 {
		t.Fatalf("Second bid incorrect")
	}
}
