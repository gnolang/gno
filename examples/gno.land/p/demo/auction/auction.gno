package auction

import (
	"std"
	"strconv"
	"time"

	"gno.land/p/demo/ownable"
	"gno.land/p/demo/ufmt"
)

// Auction struct
type Auction struct {
	Owner       ownable.Ownable // Embeds Ownable to handle ownership
	Title       string
	Description string
	Begin       time.Time
	End         time.Time
	StartingBid std.Coin
	Bids        []*Bid
}

// Bid struct
type Bid struct {
	Bidder std.Address
	Amount std.Coin
}
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
		Owner:       *ownable.NewWithAddress(owner), // Initialize Ownable with the owner
		Title:       title,
		Description: description,
		Begin:       begin,
		End:         end,
		StartingBid: startingBid,
		Bids:        []*Bid{},
	}
}

// AddBid adds a new bid to the auction
func (a *Auction) AddBid(bidder std.Address, amount std.Coin) error {
	if time.Now().Before(a.Begin) {
		return ufmt.Errorf("auction: AddBid: auction has not started yet")
	}
	if time.Now().After(a.End) {
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

// EndAuction ends the auction
func (a *Auction) EndAuction() error {
	if err := a.Owner.CallerIsOwner(); err != nil {
		return err
	}
	if time.Now().Before(a.End) {
		return ufmt.Errorf("auction: EndAuction: auction cannot end before the end time")
	}
	
	highestBid := a.Price

	if len(a.Bids) > 0 {
		winner := a.Bids[len(a.Bids)-1].Bidder
		std.Emit("AuctionEnded", "winner", winner.String(), "amount", strconv.Itoa(int(highestBid)))
	} else {
		std.Emit("AuctionEndedNoBids", "auction", a.Title)
	}

	return nil
}
