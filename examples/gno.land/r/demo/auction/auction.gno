package auction

import (
	"bytes"
	"std"
	"strconv"
	"time"

	auctionp "gno.land/p/demo/auction"
	"gno.land/p/demo/mux"
)

var (
	auctionList []*auctionp.Auction
	currentTime time.Time
	router      = mux.NewRouter()
)

var router = mux.NewRouter()

func init() {
	router.HandleFunc("", renderHomepage)
	router.HandleFunc("/auction/{id}", renderAuctionDetails)
}

// CreateAuction handles the creation of a new auction
func CreateAuction(title, description string, begin, end int64, price std.Coin) string {
	owner := std.GetOrigCaller()
	beginTime := time.Unix(begin, 0)
	endTime := time.Unix(end, 0)
	auction := auctionp.NewAuction(title, owner, description, beginTime, endTime, price)
	auctionList = append(auctionList, auction)
	return "Auction created successfully"
}

// PlaceBid handles placing a bid on an auction
func PlaceBid(id int, amount std.Coin) string {
	caller := std.GetOrigCaller()
	if id >= len(auctionList) || id < 0 {
		panic("Invalid auction ID")
	}
	auction := auctionList[id]
	if err := auction.AddBid(caller, amount); err != nil {
		panic(err.Error())
	}
	return "Bid placed successfully"
}

// EndAuction handles ending an auction
func EndAuction(id int) string {
	caller := std.GetOrigCaller()
	if id >= len(auctionList) || id < 0 {
		panic("Invalid auction ID")
	}
	auction := auctionList[id]
	if err := auction.EndAuction(caller); err != nil {
		panic(err.Error())
	}
	return "Auction ended successfully"
}

// Render renders the requested page
func Render(path string) string {
	return router.Render(path)
}

// renderHomepage renders the homepage with links to auctions list
func renderHomepage(res *mux.ResponseWriter, req *mux.Request) {
	var b bytes.Buffer
	b.WriteString("# Homepage\n\n")
	b.WriteString("## [AuctionList](/r/demo/auction:List)\n\n")

	res.Write(b.String())
}
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
	b.WriteString("# Auction List\n\n")

	b.WriteString("\n\n")
	b.WriteString("## " + auc.Title + "\n")
	b.WriteString("### Owner: " + auc.Ownable.Owner().String() + "\n")
	b.WriteString("### Description: " + auc.Description + "\n\n")
	b.WriteString("This auction starts on: " + auc.Begin.String() + " and ends on: " + auc.End.String() + "\n\n")
	b.WriteString(ufmt.Sprintf("### starting Price: %d\n", auc.Price))
	b.WriteString("[View Auction](/auction/" + strconv.Itoa(id) + ")\n")

	res.Write(b.String())
}
