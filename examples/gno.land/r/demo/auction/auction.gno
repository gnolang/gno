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
	auctionList    []*auctionp.Auction
	currentTime    time.Time
	router         = mux.NewRouter()
	globalDeposits = make(map[std.Address]int64)
)

func init() {
	router.HandleFunc("", renderHomepage)
	router.HandleFunc("/auction/{id}", renderAuctionDetails)
	//router.HandleFunc("auction:List/", renderAuctionList)
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
	owner := std.GetOrigCaller()
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
	// get the amount sent by the caller
	amountSent := std.NewCoin(denom, transaction.AmountOf(denom))
	// the bid amount to be placed
	bidAmount := std.NewCoin(denom, amount)

	// check if the ID is valid
	if id < 0 || id >= len(auctionList) {
		panic("Invalid auction ID")
	}
	auction := auctionList[id]

	banker := std.GetBanker(std.BankerTypeOrigSend)
	existingDeposit, exists := globalDeposits[caller]
	existingDepositCoin := std.NewCoin(denom, existingDeposit)

	if !exists || bidAmount.IsGTE(existingDepositCoin) {
		// Calculate the additional amount to deposit
		additionalAmount := bidAmount.Sub(existingDepositCoin)
		if additionalAmount.IsPositive() {

			if !amountSent.IsGTE(additionalAmount) {
				panic("Insufficient funds to place the bid")
			}
			banker.SendCoins(caller, std.CurrentRealm().Addr(), std.NewCoins(additionalAmount))
		}
		globalDeposits[caller] = existingDeposit + additionalAmount.Amount
	} else {

		banker.SendCoins(caller, std.CurrentRealm().Addr(), std.NewCoins(bidAmount))
		globalDeposits[caller] = existingDeposit + amount
	}

	// Place the bid
	if err := auction.AddBid(caller, bidAmount); err != nil {
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
		highestBid = auction.Bids[len(auction.Bids)-1].Amount.Amount
		denom = auction.Bids[len(auction.Bids)-1].Amount.Denom
		winner = auction.Bids[len(auction.Bids)-1].Bidder
	} else {
		highestBid = 0
	}

	// Handle fund transfers using the banker
	banker := std.GetBanker(std.BankerTypeRealmSend)
	pkgAddr := std.CurrentRealm().Addr()

	// Refund all losing bidders
	for bidder, deposit := range globalDeposits {
		if bidder != winner {
			banker.SendCoins(pkgAddr, bidder, std.Coins{{Amount: deposit, Denom: denom}})
			// Reset the deposit for the bidder
			delete(globalDeposits, bidder)
		}
	}

	// Transfer the highest bid to the auction owner
	if highestBid > 0 {
		banker.SendCoins(pkgAddr, auction.Owner.Owner(), std.Coins{{Amount: highestBid, Denom: denom}})
	}

	return "Auction ended successfully"
}

func Render(path string) string {
	return router.Render(path)
}

// renderHomepage renders the homepage with links to auctions list
func renderHomepage(res *mux.ResponseWriter, req *mux.Request) {
	var b bytes.Buffer
	b.WriteString("# Homepage\n\n")

	//b.WriteString("## [AuctionList](/r/demo/auction:List)\n\n")

	res.Write(b.String())
}

// renderAuctionList renders the list of all auctions
// func renderAuctionList(res *mux.ResponseWriter, req *mux.Request) {
//     var b bytes.Buffer
//     b.WriteString("# Auction List\n\n")

//     for i, auc := range auctionList {
//         b.WriteString("## " + auc.Title + "\n")
//         b.WriteString("### Owner: " + auc.Owner.Owner().String() + "\n")
//         b.WriteString("### Description: " + auc.Description + "\n\n")
//         b.WriteString("This auction starts on: " + auc.Begin.String() + " and ends on: " + auc.End.String() + "\n\n")
//         b.WriteString(ufmt.Sprintf("### starting Price: %d\n", auc.Price))
//         b.WriteString("[View Auction](/r/demo/auction:List/" + strconv.Itoa(i) + ")\n\n")
//     }

//     res.Write(b.String())
// }

// renderAuctionDetails renders the details of a specific auction
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

	b.WriteString("\n\n")
	b.WriteString("## " + auc.Title + "\n")
	b.WriteString("### Owner: " + auc.Owner.Owner().String() + "\n")
	b.WriteString("### Description: " + auc.Description + "\n\n")
	b.WriteString("This auction starts on: " + auc.Begin.String() + " and ends on: " + auc.End.String() + "\n\n")
	b.WriteString(ufmt.Sprintf("### starting Price: %d\n", auc.Price))
	b.WriteString("[View Auction](r/demo/auction/" + strconv.Itoa(id) + ")\n")

	res.Write(b.String())
}
