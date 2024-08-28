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
	StartingBid std.Coin
	Bids        []*Bid
	State       string // "upcoming", "ongoing", or "closed"
}

// Bid struct
type Bid struct {
	Bidder std.Address
	Amount std.Coin
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
	startingBid std.Coin,
) *Auction {
	return &Auction{
		Title:       title,
		Description: description,
		Owner:       owner,
		Begin:       begin,
		End:         end,
		StartingBid: startingBid,
		Bids:        []*Bid{},
		State:       "upcoming",
	}
}

// AddBid adds a new bid to the auction
func (a *Auction) AddBid(bidder std.Address, amount std.Coin) error {
	if now().Before(a.Begin) {
		return ufmt.Errorf("auction: AddBid: auction has not started yet")
	}
	if now().After(a.End) {
		return ufmt.Errorf("auction: AddBid: auction has already ended")
	}
	if !amount.IsGTE(a.StartingBid) {
		return ufmt.Errorf("auction: AddBid: bid amount must be higher than the current highest bid")
	}
	bid := &Bid{Bidder: bidder, Amount: amount}
	a.Bids = append(a.Bids, bid)
	a.StartingBid = amount
	std.Emit("BidPlaced", "auction", a.Title, "bidder", bidder.String(), "amount", amount.String())
	return nil
}

// IsOwnedBy checks if the given address is the owner of the auction
func (a *Auction) IsOwnedBy(address std.Address) bool {
	return a.Owner == address
}

// EndAuction ends the auction
func (a *Auction) EndAuction(caller std.Address) error {
	if caller != a.Owner {
		return ufmt.Errorf("auction: EndAuction: only the owner can end the auction")
	}
	if now().Before(a.End) {
		return ufmt.Errorf("auction: EndAuction: auction cannot end before the end time")
	}
	if a.State == "closed" {
		return ufmt.Errorf("auction: EndAuction: auction is already closed")
	}
	if len(a.Bids) == 0 {
		std.Emit("AuctionEndedNoBids", "auction", a.Title)
		return ufmt.Errorf("auction: EndAuction: auction ended with no bids")
	}
	a.State = "closed"
	highestBid := a.StartingBid
	winner := a.Bids[len(a.Bids)-1].Bidder
	std.Emit("AuctionEnded", "winner", winner.String(), "amount", highestBid.String())
	return nil
}

// SetCurrentTime sets the current time for testing purposes
func SetCurrentTime(t time.Time) {
	currentTime = t
}

// now returns the current time, allowing for mocking in tests
func now() time.Time {
	if !currentTime.IsZero() {
		return currentTime
	}
	return time.Now()
}
