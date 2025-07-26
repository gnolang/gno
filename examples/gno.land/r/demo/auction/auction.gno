package auction

import (
	"bytes"
	"strconv"
	"time"

	"std"

	auctionp "gno.land/p/demo/auction"
	"gno.land/p/demo/mux"
	"gno.land/p/demo/ufmt"
)

var (
	auctionList []*auctionp.Auction
	currentTime time.Time
	router      = mux.NewRouter()
)

const (
	RealmURL          = "/r/demo/auction"
	AuctionDetailsURL = RealmURL + ":auction/%d"
)

func init() {
	router.HandleFunc("", renderHomepage)
	router.HandleFunc("auction/{id}", renderAuctionDetails)
	//router.HandleFunc("auctionList/", renderAuctionList)
}

// CreateAuction handles the creation of a new auction
func CreateAuction(title, description string, begin, end int64, amount int64, denom string) string {
	if title == "" || description == "" {
		panic("Auction creation failed: title and description cannot be empty")
	}
	if begin >= end {
		panic("Auction creation failed: begin time must be before end time")
	}
	if amount <= 0 {
		panic("Auction creation failed: starting price must be positive")
	}
	startingPrice := std.NewCoin(denom, amount)
	owner := std.PrevRealm().Addr()
	beginTime := time.Unix(begin, 0)
	endTime := time.Unix(end, 0)
	auction := auctionp.NewAuction(title, owner, description, beginTime, endTime, startingPrice)
	auctionList = append(auctionList, auction)
	return "Auction created successfully."
}

// PlaceBid handles placing a bid on an auction
// id: the ID of the auction
// amount: the amount of the bid
// denom: the denomination of the bid amount
func PlaceBid(id int, amount int64, denom string) string {
	if amount <= 0 {
		panic("Amount must be positive")
	}

	caller := std.GetOrigCaller()
	transaction := std.GetOrigSend()
	// Amount sent by the caller
	amountSent := transaction.AmountOf(denom)
	bidAmount := amount

	// Validate auction ID
	if id < 0 || id >= len(auctionList) {
		panic("Invalid auction ID")
	}
	auction := auctionList[id]

	// Retrieve the current deposit of the caller using Banker
	banker := std.GetBanker(std.BankerTypeReadonly)
	ufmt.Println("PrevRealm Coins:", banker.GetCoins(std.PrevRealm().Addr()))
	existingCoins := banker.GetCoins(caller)
	existingDeposit := existingCoins.AmountOf(denom)

	// Calculate the additional amount needed
	additionalAmount := bidAmount - existingDeposit

	// Check if the sent amount covers the additional amount
	if additionalAmount > 0 {
		if amountSent < additionalAmount {
			panic("Insufficient funds to place the bid")
		}

	}

	// Place the bid in the auction
	if err := auction.AddBid(caller, std.NewCoin(denom, bidAmount)); err != nil {
		panic(err.Error())
	}
	return "Bid placed successfully"
}

// EndAuction ends an auction identified by the given ID, determines the winner and highest bid,
// handles fund transfers, and refunds losing bidders.
func EndAuction(id int) string {
	if id < 0 || id >= len(auctionList) {
		panic("Invalid auction ID")
	}
	auction := auctionList[id]

	// Call EndAuction to close the auction and get the winner and highest bid
	if err := auction.EndAuction(); err != nil {
		panic(err.Error())
	}

	// Determine the winner and highest bid
	var winner std.Address
	var highestBid int64
	var denom string

	if len(auction.Bids) > 0 {
		lastBid := auction.Bids[len(auction.Bids)-1]
		highestBid = lastBid.Amount.Amount
		denom = lastBid.Amount.Denom
		winner = lastBid.Bidder
	}

	// Use Banker to handle fund transfers
	banker := std.GetBanker(std.BankerTypeRealmSend)
	pkgAddr := std.CurrentRealm().Addr()

	// Refund all losing bidders
	for _, bid := range auction.Bids {
		bidder := bid.Bidder
		if bidder != winner {
			// Get the bidder's deposit directly
			bidderDeposit := banker.GetCoins(bidder).AmountOf(denom)

			// Refund the bidder
			if bidderDeposit > 0 {
				banker.SendCoins(pkgAddr, bidder, std.Coins{{Denom: denom, Amount: bidderDeposit}})
			}
		}
	}

	// Transfer the highest bid to the auction owner
	if highestBid > 0 {
		banker.SendCoins(pkgAddr, auction.Owner.Owner(), std.Coins{{Denom: denom, Amount: highestBid}})
	}

	return "Auction ended successfully"
}

// renderHomepage renders the homepage with links to auctions list
func renderHomepage(res *mux.ResponseWriter, req *mux.Request) {
	var b bytes.Buffer
	b.WriteString("# Homepage\n\n")
	b.WriteString("## Auction List\n\n")

	for i, auc := range auctionList {
		link := ufmt.Sprintf(AuctionDetailsURL, i)
		b.WriteString("### [" + auc.Title + "](" + link + ")\n\n")
	}

	res.Write(b.String())
}

func renderAuctionDetails(res *mux.ResponseWriter, req *mux.Request) {
	idStr := req.GetVar("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id >= len(auctionList) || id < 0 {
		res.Write("Invalid auction ID")
		return
	}

	auc := auctionList[id]
	var b bytes.Buffer
	b.WriteString("# Auction Details\n\n")
	b.WriteString("## " + auc.Title + "\n")
	b.WriteString("### Owner: " + auc.Owner.Owner().String() + "\n")
	b.WriteString("### Description: " + auc.Description + "\n\n")
	b.WriteString("Begin : " + auc.Begin.String() + "\n")
	b.WriteString("End : " + auc.End.String() + "\n")
	b.WriteString("Price : " + auc.Price.String() + "\n")

	res.Write(b.String())
}

func Render(path string) string {
	return router.Render(path)
}
