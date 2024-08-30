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
	// router      *mux.Router
)

var router = mux.NewRouter()

func init() {
	router.HandleFunc("", renderHomepage)
	router.HandleFunc("create/", renderCreateAuction)
	router.HandleFunc("upcoming/", renderUpcomingAuctions)
	router.HandleFunc("ongoing/", renderOngoingAuctions)
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

// renderHomepage renders the homepage with links to different sections
func renderHomepage(res *mux.ResponseWriter, req *mux.Request) {
	var b bytes.Buffer
	b.WriteString("<h1><center>Auction Home</center></h1>\n\n")
	b.WriteString("<h2><a href='/r/demo/auction:create/'>Create New Auction</a></h2>\n\n")
	b.WriteString("<h2><a href='/r/demo/auction:upcoming/'>Upcoming Auctions</a></h2>\n\n")
	b.WriteString("<h2><a href='/r/demo/auction:ongoing/'>Ongoing Auctions</a></h2>\n\n")
	res.Write(b.String())
}

// renderCreateAuction renders the create auction page
func renderCreateAuction(res *mux.ResponseWriter, req *mux.Request) {
	var b bytes.Buffer
	b.WriteString("<html><body>")

	b.WriteString("#Create New Auction\n")
	b.WriteString("<form action='/create' method='post'>\n")
	b.WriteString("Title: <input type='text' name='title'><br>\n")
	b.WriteString("Description: <input type='text' name='description'><br>\n")
	b.WriteString("Begin: <input type='datetime-local' name='begin'><br>\n")
	b.WriteString("End: <input type='datetime-local' name='end'><br>\n")
	b.WriteString("Starting Price: <input type='number' name='price'><br>\n")
	b.WriteString("<input type='submit' value='Create Auction'>\n")
	b.WriteString("</form>\n")
	b.WriteString("</body></html>")
	res.Write(b.String())
}

// renderUpcomingAuctions renders the upcoming auctions page
func renderUpcomingAuctions(res *mux.ResponseWriter, req *mux.Request) {
	var b bytes.Buffer
	b.WriteString("<h1>Upcoming Auctions</h1>\n")

	for i, auc := range auctionList {
		if auc.State == "upcoming" {
			b.WriteString("\n\n")
			b.WriteString("## " + auc.Title)
			b.WriteString("\n\n")
			b.WriteString("### Owner: " + auc.Owner.String() + "\n")
			b.WriteString("### Description: " + auc.Description + "\n\n")
			b.WriteString("This auction starts on: " + auc.Begin.String() + " and ends on: " + auc.End.String() + "\n\n")
			b.WriteString("### Starting Price: " + strconv.FormatUint(auc.Price, 10) + "\n")
			b.WriteString("[View Auction](/auction/" + strconv.Itoa(i) + ")\n")
		}
	}
	res.Write(b.String())
}

// renderOngoingAuctions renders the ongoing auctions page
func renderOngoingAuctions(res *mux.ResponseWriter, req *mux.Request) {
	var b bytes.Buffer
	b.WriteString("<h1>Ongoing Auctions</h1>\n")

	for i, auc := range auctionList {
		if auc.State == "ongoing" {
			b.WriteString("##" + auc.Title + "\n")
			b.WriteString("### Owner: " + auc.Owner.String() + "\n")
			b.WriteString("### Description: " + auc.Description + "\n\n")
			b.WriteString("This auction started on: " + auc.Begin.String() + " and ends on: " + auc.End.String() + "\n\n")
			b.WriteString("### Current Price: " + strconv.FormatUint(auc.Price, 10) + "\n")
			b.WriteString("[View Auction](/auction/" + strconv.Itoa(i) + ")\n")
		}
	}
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
	b.WriteString("<h1>Auction Details</h1>\n")
	b.WriteString("## " + auc.Title + "\n")
	b.WriteString("### Owner: " + auc.Owner.String() + "\n")
	b.WriteString("### Description: " + auc.Description + "\n\n")
	b.WriteString("This auction starts on: " + auc.Begin.String() + " and ends on: " + auc.End.String() + "\n\n")
	b.WriteString("### Current Price: " + strconv.FormatUint(auc.Price, 10) + "\n")

	if auc.State == "ongoing" {
		b.WriteString("<form action='/bid' method='post'>\n")
		b.WriteString("Amount: <input type='number' name='amount'><br>\n")
		b.WriteString("<input type='hidden' name='id' value='" + strconv.Itoa(id) + "'>\n")
		b.WriteString("<input type='submit' value='Place Bid'>\n")
		b.WriteString("</form>\n")
	}

	res.Write(b.String())
}
