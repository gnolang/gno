package treasury

import (
	"chain/runtime"
	"net/url"
	"strconv"
	"strings"

	"gno.land/p/moul/md"
	"gno.land/p/moul/mdtable"
	"gno.land/p/nt/avl/v0/pager"
	"gno.land/p/nt/mux/v0"
	"gno.land/p/nt/ufmt/v0"
)

const (
	DefaultHistoryPreviewSize = 5  // Number of payments in the history preview.
	DefaultHistoryPageSize    = 20 // Number of payments per page in the history.
)

// Render renders content based on the given path.
func (t *Treasury) Render(path string) string {
	return t.router.Render(path)
}

// RenderLanding renders the landing page of the treasury.
func (t *Treasury) RenderLanding(path string) string {
	var out string

	// Render each banker.
	for _, bankerID := range t.ListBankerIDs() {
		out += t.RenderBanker(bankerID, path)
	}

	return out
}

// RenderBanker renders the details of a specific banker.
func (t *Treasury) RenderBanker(bankerID string, path string) string {
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

	historySize := DefaultHistoryPreviewSize

	// Check if the query parameter "history_size" is present and parse it.
	if req, err := url.Parse(path); err == nil && req.Query() != nil {
		size, err := strconv.Atoi(req.Query().Get("history_size"))
		if err == nil && size >= 0 {
			historySize = size
		}
	}

	// Skip history rendering if historySize is 0.
	if historySize == 0 {
		return out
	}

	// Render history section.
	out += md.H3("History")
	history, _ := t.History(bankerID, 1, historySize)
	if len(history) == 0 {
		out += md.Paragraph("No payments sent yet.")
	} else {
		if len(history) == 1 {
			out += md.Paragraph("Last payment:")
		} else {
			count := strconv.FormatInt(int64(len(history)), 10)
			out += md.Paragraph("Last " + count + " payments:")
		}

		// Render each payment in the history.
		for _, payment := range history {
			out += md.BulletItem(payment.String())
		}
		out += "\n"

		// Get the current Realm's package path for linking.
		curRealm := runtime.CurrentRealm().PkgPath()
		if from := strings.IndexRune(curRealm, '/'); from >= 0 {
			out += md.Link(
				"See full history",
				ufmt.Sprintf("%s:%s/history", curRealm[from:], bankerID),
			)
		}
	}

	return out
}

// RenderBankerHistory renders the payment history of a specific banker.
func (t *Treasury) RenderBankerHistory(bankerID string, path string) string {
	// Get the banker record corresponding to this ID if it exists.
	br, ok := t.bankers.Get(bankerID)
	if !ok {
		return md.Paragraph("Banker not found: " + bankerID)
	}
	history := br.(*bankerRecord).history

	// Render banker history title.
	out := md.H2(bankerID + " Banker History")

	// Get the current page of tokens based on the request path.
	p := pager.NewPager(history.Tree(), DefaultHistoryPageSize, true)
	page, err := p.GetPageByPath(path)
	if err != nil {
		return md.Paragraph("Error retrieving page: " + err.Error())
	}

	// Render full history section.
	if history.Len() == 0 {
		out += md.Paragraph("No payments sent yet.")
	} else {
		if history.Len() == 1 {
			out += md.Paragraph("1 payment:")
		} else {
			count := strconv.FormatInt(int64(history.Len()), 10)
			out += md.Paragraph(count + " payments (sorted by latest, descending):")
		}
		for _, item := range page.Items {
			out += md.BulletItem(item.Value.(Payment).String())
		}
	}
	out += "\n"

	// Add the page picker.
	out += md.Paragraph(page.Picker(path))

	return out
}

// initRenderRouter registers the routes for rendering the treasury pages.
func (t *Treasury) initRenderRouter() {
	t.router = mux.NewRouter()

	// Landing page.
	t.router.HandleFunc("", func(res *mux.ResponseWriter, req *mux.Request) {
		res.Write(t.RenderLanding(req.RawPath))
	})

	// Banker details.
	t.router.HandleFunc("{banker}", func(res *mux.ResponseWriter, req *mux.Request) {
		res.Write(t.RenderBanker(req.GetVar("banker"), req.RawPath))
	})

	// Banker full history.
	t.router.HandleFunc("{banker}/history", func(res *mux.ResponseWriter, req *mux.Request) {
		res.Write(t.RenderBankerHistory(req.GetVar("banker"), req.RawPath))
	})
}
