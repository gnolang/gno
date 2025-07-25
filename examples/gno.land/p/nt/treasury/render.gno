package treasury

import (
	"net/url"
	"strconv"

	"gno.land/p/demo/mux"
	"gno.land/p/moul/md"
	"gno.land/p/moul/mdtable"
)

const DefaultHistorySize = 5 // Number of last payments to show in the history.

// Render renders content based on the given path.
func (t *Treasury) Render(path string) string {
	return t.router.Render(path)
}

// RenderLanding renders the landing page of the treasury.
func (t *Treasury) RenderLanding(historySize uint) string {
	var out string

	// Render each banker.
	for _, bankerID := range t.ListBankerIDs() {
		out += t.RenderBanker(bankerID, historySize)
	}

	return out
}

// RenderBanker renders the details of a specific banker.
func (t *Treasury) RenderBanker(bankerID string, historySize uint) string {
	// Get the banker associated to this ID.
	br, ok := t.bankers.Get(bankerID)
	if !ok {
		return md.Paragraph("Banker not found: " + bankerID)
	}
	banker := br.(*bankerRecord).banker

	// Render banker title.
	out := md.H2(bankerID + " Banker")

	// Render address section.
	out += md.H3("Address")
	out += md.Paragraph(banker.Address())

	// Render balances section.
	out += md.H3("Balances")
	balances := banker.Balances()
	if len(balances) == 0 {
		out += md.Paragraph("No balances found.")
	} else {
		table := mdtable.Table{Headers: []string{"Denom", "Amount"}}
		for _, balance := range balances {
			table.Append([]string{balance.Denom, strconv.FormatInt(balance.Amount, 10)})
		}
		out += table.String()
	}

	// Skip history rendering if historySize is 0.
	if historySize == 0 {
		return out
	}

	// Render history section.
	out += md.H3("History")
	history, _ := t.History(bankerID, historySize)
	if len(history) == 0 {
		out += md.Paragraph("No payments sent yet.")
	} else {
		if len(history) == 1 {
			out += md.Paragraph("Last payment:")
		} else {
			count := strconv.FormatInt(int64(len(history)), 10)
			out += md.Paragraph("Last " + count + " payments:")
		}
		for _, payment := range history {
			out += md.BulletItem(payment.String())
		}
		out += md.Link(
			"See full history",
			"/r/gov/dao/v3/treasury:"+bankerID+"/history",
		)
	}

	return out
}

// RenderBankerHistory renders the payment history of a specific banker.
func (t *Treasury) RenderBankerHistory(bankerID string) string {
	// Check if the banker exists.
	if !t.HasBanker(bankerID) {
		return md.Paragraph("Banker not found: " + bankerID)
	}

	// Render banker history title.
	out := md.H2(bankerID + " Banker History")

	// Render full history section.
	history, _ := t.History(bankerID, 0) // Get full history.
	if len(history) == 0 {
		out += md.Paragraph("No payments sent yet.")
	} else {
		if len(history) == 1 {
			out += md.Paragraph("1 payment:")
		} else {
			count := strconv.FormatInt(int64(len(history)), 10)
			out += md.Paragraph(count + " payments (sorted by latest, descending):")
		}
		for _, payment := range history {
			out += md.BulletItem(payment.String())
		}
	}

	return out
}

// initRenderRouter registers the routes for rendering the treasury pages.
func (t *Treasury) initRenderRouter() {
	t.router = mux.NewRouter()

	// getHistorySize retrieves the history size from the query parameters.
	getHistorySize := func(query url.Values) uint {
		// Check if the query parameter "history_size" is present and parse it.
		if size, err := strconv.Atoi(query.Get("history_size")); err == nil && size >= 0 {
			return uint(size)
		}

		// If not present, return the default history size.
		return DefaultHistorySize
	}

	// Landing page.
	t.router.HandleFunc("", func(res *mux.ResponseWriter, req *mux.Request) {
		historySize := getHistorySize(req.Query)
		res.Write(t.RenderLanding(historySize))
	})

	// Banker details.
	t.router.HandleFunc("{banker}", func(res *mux.ResponseWriter, req *mux.Request) {
		historySize := getHistorySize(req.Query)
		res.Write(t.RenderBanker(req.GetVar("banker"), historySize))
	})

	// Banker full history.
	t.router.HandleFunc("{banker}/history", func(res *mux.ResponseWriter, req *mux.Request) {
		res.Write(t.RenderBankerHistory(req.GetVar("banker")))
	})
}
