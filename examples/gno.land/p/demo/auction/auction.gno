package auction //package auction provides a simple auction system

import (
	"std"
	"time"

	"gno.land/p/demo/ownable"
	"gno.land/p/demo/ufmt"
)

// Auction represents an auction with various attributes:
//
//	Owner: The owner of the auction, handled by embedding the Ownable.
//	Title: The title of the auction.
//	Description: A brief description of the auction.
//	Begin: The start time of the auction.
//	End: The end time of the auction.
//	Price: The starting price of the auction, represented by the amount and the denomination.
//	Bids: A list of bids placed on the auction, represented as pointers to Bid.
type Auction struct {
	Owner       ownable.Ownable
	Title       string
	Description string
	Begin       time.Time
	End         time.Time
	Price       std.Coin
	Bids        []*Bid
}

// Bid represents a bid in the auction with a bidder's address and the bid amount and denom.
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
	price std.Coin,
) *Auction {
	return &Auction{
		Owner:       *ownable.NewWithAddress(owner),
		Title:       title,
		Description: description,
		Begin:       begin,
		End:         end,
		Price:       price,
		Bids:        []*Bid{},
	}
}

// auctionIsActive checks if the auction is currently active.
func (a *Auction) auctionIsActive() error {
	now := time.Now()
	if now.Before(a.Begin) {
		return ufmt.Errorf("auctionIsActive: auction has not started yet")
	}
	if now.After(a.End) {
		return ufmt.Errorf("auctionIsActive: auction has already ended")
	}
	return nil
}

// validateBidAmount checks if the provided bid amount is valid.
func (a *Auction) validateBidAmount(amount std.Coin) error {
	if !amount.IsPositive() {
		return ufmt.Errorf("AddBid: bid amount must be positive")
	}
	if len(a.Bids) > 0 {
		lastBid := a.Bids[len(a.Bids)-1]
		if !amount.IsGTE(lastBid.Amount) {
			return ufmt.Errorf("AddBid: bid amount must be higher than the last bid")
		}
	} else {
		if !amount.IsGTE(a.Price) {
			return ufmt.Errorf("AddBid: bid amount must be higher than the starting price")
		}
	}
	return nil
}

// placeBid allows a bidder to place a bid on the auction.
// It appends the new bid to the list of bids and updates the auction's price.
func (a *Auction) placeBid(bidder std.Address, amount std.Coin) {
	bid := &Bid{Bidder: bidder, Amount: amount}
	a.Bids = append(a.Bids, bid)
	a.Price = amount
	std.Emit(
		"BidPlaced",
		"auction",
		a.Title,
		"bidder",
		bidder.String(),
		"amount",
		amount.String(),
	)
}

// AddBid adds a new bid to the auction and returns an error if the bid is invalid
func (a *Auction) AddBid(bidder std.Address, amount std.Coin) error {
	if err := a.auctionIsActive(); err != nil {
		return err
	}
	if err := a.validateBidAmount(amount); err != nil {
		return err
	}

	a.placeBid(bidder, amount)
	return nil
}

// EndAuction ends the auction and emits an event with the winner and the highest bid amount.
// If there are no bids, it emits an event indicating no bids were placed.
func (a *Auction) EndAuction() error {
	if err := a.Owner.CallerIsOwner(); err != nil {
		return err
	}
	if time.Now().Before(a.End) {
		return ufmt.Errorf("EndAuction: auction cannot end before the end time")
	}

	if len(a.Bids) == 0 {
		std.Emit(
			"AuctionEnded",
			"auction",
			a.Title,
			"noBids",
		)
		return nil
	}

	highestBid := a.Bids[len(a.Bids)-1].Amount
	winner := a.Bids[len(a.Bids)-1].Bidder
	std.Emit(
		"AuctionEnded",
		"auction",
		a.Title,
		"winner",
		winner.String(),
		"amount",
		highestBid.String(),
	)

	return nil
}
