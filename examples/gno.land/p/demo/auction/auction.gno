package auction

import (
	"std"
	"strconv"
	"time"

	"gno.land/p/demo/avl"
)

// Main struct
type Auction struct {
	Title       string
	Owner       std.Address
	Description string
	Begin       time.Time
	End         time.Time
	Price       uint64
	Bids        *avl.Tree // key: std.Address, value: Bid
}

type Bid struct {
	Bidder std.Address
	Amount uint64
}

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
		Bids:        avl.NewTree(),
		Begin:       begin,
		End:         end,
		Price:       minPrice,
	}
}

func (A Auction) GetTitle() string {
	return A.Title
}

func (A Auction) GetDescription() string {
	return A.Description
}

func (A Auction) GetOwner() std.Address {
	return A.Owner
}

func (A Auction) GetBegin() time.Time {
	return A.Begin
}

func (A Auction) GetEnd() time.Time {
	return A.End
}

func (A Auction) GetPrice() uint64 {
	return A.Price
}

func (A Auction) GetBids() []*Bid {
	bids := make([]*Bid, 0, A.Bids.Size())
	A.Bids.Iterate("", "", func(key string, value interface{}) bool {
		bids = append(bids, value.(*Bid))
		return false
	})
	return bids
}

func (A Auction) IsOwner(address std.Address) bool {
	return address == A.GetOwner()
}

func NewBid(bidder std.Address, amount uint64) *Bid {
	return &Bid{
		Bidder: bidder,
		Amount: amount,
	}
}

func (A *Auction) AddBid(bidder std.Address, amount uint64) {
	key := strconv.Itoa(A.Bids.Size())
	bid := NewBid(bidder, amount)
	A.Bids.Set(key, bid)
	A.Price = amount
}

func (B Bid) GetBidder() std.Address {
	return B.Bidder
}

func (B Bid) GetAmount() uint64 {
	return B.Amount
}
