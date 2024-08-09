package auction

import (
	"bytes"
	"std"
	"strconv"
	"time"

	auctionp "gno.land/p/demo/auction"
	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
)

var (
	auctionTree *avl.Tree
	currentTime time.Time
)

// Initialize the realm with auction tree
func init() {
	auctionTree = avl.NewTree()
}

// Mockable function to get current time
func now() time.Time {
	if !currentTime.IsZero() {
		return currentTime
	}
	return time.Now()
}

// NewAuction creates a new auction in the realm
func NewAuction(
	title string,
	description string,
	begin int64,
	end int64,
	price uint64,
) *auctionp.Auction {
	txSender := std.GetOrigCaller()

	if begin < now().Unix() {
		panic("begin has to be in the future")
	}
	if end <= now().Unix() {
		panic("end has to be in the future")
	}
	if end <= begin {
		panic("end has to be after begin")
	}
	if price <= 0 {
		panic("price has to be positive")
	}

	auc := auctionp.NewAuction(
		title,
		txSender,
		description,
		time.Unix(begin, 0),
		time.Unix(end, 0),
		price,
	)
	id := strconv.Itoa(auctionTree.Size())
	auctionTree.Set(id, auc)
	return auc
}

// AddBid places a bid in the auction
func AddBid(auctionID string, price uint64) string {
	bidder := std.GetOrigCaller()
	auc, ok := auctionTree.Get(auctionID)
	if !ok {
		panic("auction does not exist")
	}
	auction := auc.(*auctionp.Auction)
	if auction.GetBegin().After(now()) {
		panic("auction is not open yet")
	}
	if auction.GetEnd().Before(now()) {
		panic("auction is closed")
	}
	if auction.IsOwner(bidder) {
		panic("owner cannot bid")
	}
	if auction.GetPrice() >= price {
		panic("price has to be higher than minimum price")
	}

	auction.AddBid(bidder, price)
	return "bid placed"
}

// Render renders the state of the realm
func Render(path string) string {
	if path == "" {
		return renderHomepage()
	}
	return "unknown page"
}

// renderHomepage renders the homepage of the realm
func renderHomepage() string {
	var b bytes.Buffer
	b.WriteString("<h1><center>Auctions</center></h1>\n\n")
	if auctionTree.Size() == 0 {
		b.WriteString("## No auctions available\n")
		return b.String()
	}
	auctionTree.Iterate("", "", func(key string, value interface{}) bool {
		auc := value.(*auctionp.Auction)
		b.WriteString("## " + auc.GetTitle() + "\n")
		b.WriteString("### Owner: " + auc.GetOwner().String() + "\n")
		b.WriteString("### Description: " + auc.GetDescription() + "\n\n")
		b.WriteString("This auction starts on: " + auc.GetBegin().String() + " and ends on: " + auc.GetEnd().String() + "\n\n")
		b.WriteString(ufmt.Sprintf("### Bids %d \n", auc.GetPrice()))
		b.WriteString("## Bids\n")
		bids := auc.GetBids()
		for _, bid := range bids {
			b.WriteString("Bidder: " + bid.GetBidder().String() + ", Amount: " + strconv.FormatUint(bid.GetAmount(), 10) + "\n")
		}
		return false
	})
	return b.String()
}

// Helper function to set the mock current time
func setcurrentTime(t time.Time) {
	currentTime = t
}

// Helper function to reset the mock current time
func resetcurrentTime() {
	currentTime = time.Time{}
}
