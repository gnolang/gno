package auction

import (
	"std"
	"strconv"
	"time"

	"gno.land/p/demo/avl"
)

// Main struct
type Auction struct {
	title       string
	owner       std.Address
	description string
	begin       time.Time
	end         time.Time
	price       uint64
	bids        *avl.Tree // key: std.Address, value: Bid
}

type Bid struct {
	bidder std.Address
	amount uint64
}

type EventEnd struct {
	winner std.Address
	amount uint64
}

func NewAuction(
	title string,
	owner std.Address,
	description string,
	begin time.Time,
	end time.Time,
	minPrice uint64,
) *Auction {
	auction := &Auction{
		title:       title,
		description: description,
		owner:       owner,
		bids:        avl.NewTree(),
		begin:       begin,
		end:         end,
		price:       minPrice,
	}
	std.Emit("title", title, "time", begin.String())
	return auction
}

// GetTitle returns the title of the auction.
func (a Auction) GetTitle() string {
	return a.title
}

// GetDescription returns the description of the auction.
func (a Auction) GetDescription() string {
	return a.description
}

// GetOwner returns the address of the auction owner.
func (a Auction) GetOwner() std.Address {
	return a.owner
}

// GetBegin returns the start time of the auction.
func (a Auction) GetBegin() time.Time {
	return a.begin
}

// GetEnd returns the end time of the auction.
func (a Auction) GetEnd() time.Time {
	return a.end
}

// GetPrice returns the current highest bid amount.
func (a Auction) GetPrice() uint64 {
	return a.price
}

// GetBids returns an array of all the bids made in the auction.
func (a Auction) GetBids() []*Bid {
	bids := make([]*Bid, 0, a.bids.Size())
	a.bids.Iterate("", "", func(key string, value interface{}) bool {
		bids = append(bids, value.(*Bid))
		return false
	})
	return bids
}

// IsOwner checks if the given address is the owner of the auction.
func (a Auction) IsOwner(address std.Address) bool {
	return address == a.GetOwner()
}

// NewBid creates a new Bid instance.
func NewBid(bidder std.Address, amount uint64) *Bid {
	return &Bid{
		bidder: bidder,
		amount: amount,
	}
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
