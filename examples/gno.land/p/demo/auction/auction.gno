package auction

import (
	"std"
	"strconv"
	"time"

	"gno.land/p/demo/ufmt"
)

// Auction struct
type Auction struct {
	Title       string
	Owner       std.Address
	Description string
	Begin       time.Time
	End         time.Time
	Price       uint64
	Bids        []*Bid
	State       string // "upcoming", "ongoing", or "closed"
}

// Bid struct
type Bid struct {
	Bidder std.Address
	Amount uint64
}

var (
	auctionList []*Auction
	currentTime time.Time
)

// NewAuction creates a new auction
func NewAuction(
	title string,
	owner std.Address,
	description string,
	begin time.Time,
	end time.Time,
	minPrice uint64,
) *Auction {
	return &Auction{
		Title:       title,
		Description: description,
		Owner:       owner,
		Begin:       begin,
		End:         end,
		Price:       minPrice,
		Bids:        []*Bid{},
		State:       "upcoming",
	}
}

// IsOwner checks if the given address is the owner of the auction.
func (a Auction) IsOwnedBy(address std.Address) bool {
	return address == a.GetOwner()
}

// AddBid adds a new bid to the auction.
func (a *Auction) AddBid(bidder std.Address, amount uint64) {
	if amount <= a.price {
		panic("bid amount must be higher than current highest bid")
	}
	key := strconv.Itoa(a.bids.Size())
	bid := NewBid(bidder, amount)
	a.bids.Set(key, bid)
	a.price = amount
	std.Emit("BidPlaced", "bidder", bidder.String(), "amount", strconv.FormatUint(amount, 10))
}

// GetBidder returns the address of the bidder.
func (b Bid) GetBidder() std.Address {
	return b.bidder
}

// GetAmount returns the amount of the bid.
func (b Bid) GetAmount() uint64 {
	return b.amount
}

// EndAuction ends the auction and emits the AuctionEnded event.
func (a *Auction) EndAuction() {
	if time.Now().Before(a.end) {
		panic("Auction cannot end before the end time")
	}
	if a.bids.Size() == 0 {
		panic("No bids placed")
	}
	highestBid := a.GetPrice()
	winner := a.GetBids()[a.bids.Size()-1].GetBidder()
	std.Emit("AuctionEnded", "winner", winner.String(), "amount", strconv.FormatUint(highestBid, 10))
}
