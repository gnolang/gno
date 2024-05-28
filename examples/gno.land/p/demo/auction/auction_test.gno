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
