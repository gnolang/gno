package indexer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// newTestClient returns a Client pointed at a stub indexer, plus a pointer to
// the request counter so a test can assert that the breaker really skips the
// network rather than just swallowing the error.
func newTestClient(t *testing.T, handler http.HandlerFunc) (*Client, *atomic.Int32) {
	t.Helper()
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		handler(w, r)
	}))
	t.Cleanup(srv.Close)
	return New(srv.URL, ""), &calls
}

func respond(w http.ResponseWriter, body string) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = io.WriteString(w, body)
}

func TestQueryDecodesData(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		respond(w, `{"data":{"latestBlockHeight":42}}`)
	})

	got, err := c.LatestBlockHeight(context.Background())
	if err != nil {
		t.Fatalf("LatestBlockHeight: %v", err)
	}
	if got != 42 {
		t.Fatalf("height = %d, want 42", got)
	}
}

func TestLatestBlockHeightCachesWithinTTL(t *testing.T) {
	c, calls := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		respond(w, `{"data":{"latestBlockHeight":7}}`)
	})

	for range 3 {
		if _, err := c.LatestBlockHeight(context.Background()); err != nil {
			t.Fatalf("LatestBlockHeight: %v", err)
		}
	}
	if calls.Load() != 1 {
		t.Fatalf("requests = %d, want 1 (height must be cached within heightTTL)", calls.Load())
	}
}

func TestQueryClassifiesGraphQLErrors(t *testing.T) {
	tests := []struct {
		name string
		body string
		want error
	}{
		{"not found", `{"errors":[{"message":"item not found in storage"}]}`, ErrNotFound},
		{"element cap", `{"data":{"getTransactions":[]},"errors":[{"message":"max elements per query reached"}]}`, ErrTooLarge},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				respond(w, tt.body)
			})
			var out struct {
				Txs []Tx `json:"getTransactions"`
			}
			err := c.Query(context.Background(), `{ getTransactions { hash } }`, &out)
			if !errors.Is(err, tt.want) {
				t.Fatalf("error = %v, want %v", err, tt.want)
			}
		})
	}
}

// A capped response is partial, not empty: the resolver returns the rows it
// already walked alongside the error, and a DESC query's kept rows are the
// newest ones. Losing them would turn a full "recent" panel into an empty one.
func TestCappedResponseKeepsPartialRows(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		respond(w, `{"data":{"getTransactions":[{"hash":"abc"}]},"errors":[{"message":"max elements per query"}]}`)
	})

	var out struct {
		Txs []Tx `json:"getTransactions"`
	}
	err := c.Query(context.Background(), `{ getTransactions { hash } }`, &out)
	if !errors.Is(err, ErrTooLarge) {
		t.Fatalf("error = %v, want ErrTooLarge", err)
	}
	if len(out.Txs) != 1 || out.Txs[0].Hash != "abc" {
		t.Fatalf("partial rows = %+v, want one tx with hash abc", out.Txs)
	}
}

func TestNonOKStatusReportsStatus(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = io.WriteString(w, "<html>rate limited</html>")
	})

	var out struct{}
	err := c.Query(context.Background(), `{ latestBlockHeight }`, &out)
	if err == nil || !strings.Contains(err.Error(), "403") {
		t.Fatalf("error = %v, want one naming the 403 status", err)
	}
}

// An indexer that is down must stop costing a timeout per request: after
// breakerThreshold failures the client answers without reaching the network.
func TestBreakerSkipsNetworkAfterRepeatedFailures(t *testing.T) {
	c, calls := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	var out struct{}
	for range breakerThreshold {
		_ = c.Query(context.Background(), `{ latestBlockHeight }`, &out)
	}
	if calls.Load() != breakerThreshold {
		t.Fatalf("requests before breaker = %d, want %d", calls.Load(), breakerThreshold)
	}

	err := c.Query(context.Background(), `{ latestBlockHeight }`, &out)
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("error = %v, want ErrUnavailable", err)
	}
	if calls.Load() != breakerThreshold {
		t.Fatalf("requests after breaker = %d, want it to stay at %d", calls.Load(), breakerThreshold)
	}
}

// A miss describes the question, not the indexer's health. Counting misses as
// failures would let three ordinary "no such hash" lookups take the indexer
// out of service for everyone.
func TestNotFoundDoesNotOpenBreaker(t *testing.T) {
	c, calls := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		respond(w, `{"errors":[{"message":"item not found in storage"}]}`)
	})

	for range breakerThreshold + 2 {
		if _, err := c.TxByHash(context.Background(), "deadbeef"); !errors.Is(err, ErrNotFound) {
			t.Fatalf("error = %v, want ErrNotFound", err)
		}
	}
	if calls.Load() != breakerThreshold+2 {
		t.Fatalf("requests = %d, want all %d to reach the network", calls.Load(), breakerThreshold+2)
	}
}

// gqlString is the only thing standing between caller-supplied text (a search
// box, a URL segment) and the query document.
func TestGqlStringEscapes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
	}{
		{"quote and brace", `" } getTransactions(where: {`},
		{"backslash", `a\b`},
		{"newline", "a\nb"},
		{"unicode", "héllo"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := gqlString(tt.in)
			var back string
			if err := json.Unmarshal([]byte(got), &back); err != nil {
				t.Fatalf("%q is not a valid string literal: %v", got, err)
			}
			if back != tt.in {
				t.Fatalf("round trip = %q, want %q", back, tt.in)
			}
		})
	}
}

// The selection set must never ask for package file bodies: a deploy carries
// its whole source, and a page of them is megabytes on the wire.
func TestTxFieldsExcludeFileBodies(t *testing.T) {
	t.Parallel()

	if strings.Contains(txFields, "body") {
		t.Fatal("txFields selects file bodies; a page of deploys would be megabytes")
	}
}

func TestRecentWalksDisjointBands(t *testing.T) {
	var bounds [][2]int
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req gqlRequest
		_ = json.Unmarshal(body, &req)

		if strings.Contains(req.Query, "latestBlockHeight") {
			respond(w, `{"data":{"latestBlockHeight":100000}}`)
			return
		}
		bounds = append(bounds, parseBounds(t, req.Query))
		// Never enough, so the walk runs its full course.
		respond(w, `{"data":{"getTransactions":[]}}`)
	})

	if _, err := c.RecentByPackage(context.Background(), "gno.land/r/demo/boards", 5); err != nil {
		t.Fatalf("RecentByPackage: %v", err)
	}
	if len(bounds) < 2 {
		t.Fatalf("bands = %v, want at least two", bounds)
	}

	// Each band must start exactly where the previous one stopped: a window
	// restarted from the tip re-scanned everything the earlier steps covered.
	for i, b := range bounds {
		gt, lt := b[0], b[1]
		if gt >= lt {
			t.Fatalf("band %d is empty or inverted: (%d, %d)", i, gt, lt)
		}
		if i == 0 {
			continue
		}
		prevGt := bounds[i-1][0]
		if lt != prevGt+1 {
			t.Errorf("band %d starts at lt=%d, want %d — bands must be contiguous and disjoint",
				i, lt, prevGt+1)
		}
	}

	// The last band reaches the bottom of the chain and drops its lower
	// bound, because `gt` would exclude block 0 where genesis packages live.
	if last := bounds[len(bounds)-1]; last[0] != -1 {
		t.Errorf("last band still carries gt=%d, want none at the chain bottom", last[0])
	}
}

// parseBounds reads `block_height: { gt: N, lt: M }` back out of a query.
// gt is -1 when the band carries no lower bound.
func parseBounds(t *testing.T, q string) [2]int {
	t.Helper()

	gt, lt := -1, -1
	if m := regexp.MustCompile(`gt: (\d+)`).FindStringSubmatch(q); m != nil {
		gt, _ = strconv.Atoi(m[1])
	}
	m := regexp.MustCompile(`lt: (\d+)`).FindStringSubmatch(q)
	if m == nil {
		t.Fatalf("no lt bound in query:\n%s", q)
	}
	lt, _ = strconv.Atoi(m[1])
	return [2]int{gt, lt}
}

func TestRecentStopsOnceItHasEnough(t *testing.T) {
	queries := 0
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req gqlRequest
		_ = json.Unmarshal(body, &req)

		if strings.Contains(req.Query, "latestBlockHeight") {
			respond(w, `{"data":{"latestBlockHeight":100000}}`)
			return
		}
		queries++
		respond(w, `{"data":{"getTransactions":[{"hash":"abc","messages":[{"value":{"__typename":"MsgCall","pkg_path":"gno.land/r/demo/boards"}}]}]}}`)
	})

	txs, err := c.RecentByPackage(context.Background(), "gno.land/r/demo/boards", 1)
	if err != nil {
		t.Fatalf("RecentByPackage: %v", err)
	}
	if len(txs) != 1 {
		t.Fatalf("txs = %d, want 1", len(txs))
	}
	if queries != 1 {
		t.Fatalf("queries = %d, want 1 — the first band already had enough", queries)
	}
}

// The bottom window carries no lower bound at all: `gt` excludes its bound, so
// a chain launched with packages at genesis keeps them at a height no `gt`
// filter can reach.
func TestRecentDropsHeightFilterAtChainBottom(t *testing.T) {
	var lastTxQuery string
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req gqlRequest
		_ = json.Unmarshal(body, &req)

		if strings.Contains(req.Query, "latestBlockHeight") {
			respond(w, `{"data":{"latestBlockHeight":10}}`)
			return
		}
		lastTxQuery = req.Query
		respond(w, `{"data":{"getTransactions":[]}}`)
	})

	if _, err := c.RecentByPackage(context.Background(), "gno.land/r/demo/boards", 5); err != nil {
		t.Fatalf("RecentByPackage: %v", err)
	}
	// `block_height` also appears in the selection set, so assert on the
	// filter's comparator rather than on the field name.
	if strings.Contains(lastTxQuery, "gt:") {
		t.Fatalf("query at chain bottom still carries a height bound:\n%s", lastTxQuery)
	}
}

// Concurrent callers must share one tip fetch. Holding a mutex across the
// request instead would serialize every in-flight search behind a
// slow-but-healthy indexer, which the breaker never catches because nothing
// is failing.
func TestConcurrentTipFetchesCoalesce(t *testing.T) {
	release := make(chan struct{})
	c, calls := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		<-release
		respond(w, `{"data":{"latestBlockHeight":99}}`)
	})

	const callers = 8
	var wg sync.WaitGroup
	heights := make([]int, callers)
	errs := make([]error, callers)
	for i := range callers {
		wg.Go(func() {
			heights[i], errs[i] = c.LatestBlockHeight(context.Background())
		})
	}

	// Let every caller reach the client before the stub answers.
	time.Sleep(50 * time.Millisecond)
	close(release)
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Fatalf("caller %d: %v", i, err)
		}
		if heights[i] != 99 {
			t.Fatalf("caller %d: height = %d, want 99", i, heights[i])
		}
	}
	if calls.Load() != 1 {
		t.Fatalf("requests = %d, want 1 for %d concurrent callers", calls.Load(), callers)
	}
}

// The indexer is operator-configured and off-chain, so it is trusted no
// further than the chain node: an oversized body is refused before it is
// decoded, exactly as the RPC client refuses one.
func TestOversizedResponseIsRefusedBeforeDecoding(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// A syntactically valid body that never ends.
		_, _ = io.WriteString(w, `{"data":{"pad":"`)
		chunk := strings.Repeat("a", 1<<20)
		for range 10 {
			if _, err := io.WriteString(w, chunk); err != nil {
				return
			}
		}
		_, _ = io.WriteString(w, `"}}`)
	})

	var out struct{}
	err := c.Query(context.Background(), `{ latestBlockHeight }`, &out)
	if !errors.Is(err, ErrResponseTooLarge) {
		t.Fatalf("error = %v, want ErrResponseTooLarge", err)
	}
}

// A band that fails after earlier ones found rows returns those rows: they
// are the newest, and a reader is better served by them than by an error.
// They come flagged ErrPartial, since the failed band may have held more.
func TestRecentKeepsRowsWhenALaterBandFails(t *testing.T) {
	bands := 0
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req gqlRequest
		_ = json.Unmarshal(body, &req)

		if strings.Contains(req.Query, "latestBlockHeight") {
			respond(w, `{"data":{"latestBlockHeight":100000}}`)
			return
		}
		bands++
		if bands == 1 {
			respond(w, `{"data":{"getTransactions":[{"hash":"newest","messages":[{"value":{"__typename":"MsgCall","pkg_path":"gno.land/r/demo/boards"}}]}]}}`)
			return
		}
		w.WriteHeader(http.StatusBadGateway)
	})

	txs, err := c.RecentByPackage(context.Background(), "gno.land/r/demo/boards", 5)
	if !errors.Is(err, ErrPartial) {
		t.Fatalf("RecentByPackage: %v, want the first band's rows flagged ErrPartial", err)
	}
	if len(txs) != 1 || txs[0].Hash != "newest" {
		t.Fatalf("txs = %+v, want the first band's row", txs)
	}
	// The cause travels with it, for the log.
	if !strings.Contains(err.Error(), "502") {
		t.Fatalf("err = %v, want the band's failure named", err)
	}
}

// tx-indexer lets a message it could not decode through any filter. Those
// rows must neither reach the reader nor use up the walk's budget: the walk
// goes on to the next band for the real deploy.
func TestRecentDropsRowsTheFilterDidNotMatch(t *testing.T) {
	bands := 0
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req gqlRequest
		_ = json.Unmarshal(body, &req)

		if strings.Contains(req.Query, "latestBlockHeight") {
			respond(w, `{"data":{"latestBlockHeight":100000}}`)
			return
		}
		bands++
		if bands == 1 {
			respond(w, `{"data":{"getTransactions":[
				{"hash":"session","messages":[{"value":{"__typename":"UnexpectedMessage"}}]},
				{"hash":"other","messages":[{"value":{"__typename":"MsgAddPackage","package":{"path":"gno.land/r/demo/other"}}}]}
			]}}`)
			return
		}
		respond(w, `{"data":{"getTransactions":[
			{"hash":"deploy","messages":[{"value":{"__typename":"MsgAddPackage","package":{"path":"gno.land/r/demo/boards"}}}]}
		]}}`)
	})

	txs, err := c.Deploys(context.Background(), "gno.land/r/demo/boards", 1)
	if err != nil {
		t.Fatalf("Deploys: %v", err)
	}
	if len(txs) != 1 || txs[0].Hash != "deploy" {
		t.Fatalf("txs = %+v, want only the deploy of the package", txs)
	}
	if bands != 2 {
		t.Fatalf("bands = %d, want 2: rows the filter did not match must not count", bands)
	}
}

// cappedIndexer answers a band as tx-indexer does: rows DESC down to its cap,
// then an error beside the rows it kept. A filter on value alone lets a row
// through whatever it says, so each band is as dense as the chain is.
func cappedIndexer(t *testing.T, tip, maxRows int, rows []string) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req gqlRequest
		_ = json.Unmarshal(body, &req)
		if strings.Contains(req.Query, "latestBlockHeight") {
			respond(w, fmt.Sprintf(`{"data":{"latestBlockHeight":%d}}`, tip))
			return
		}
		band := parseBounds(t, req.Query)
		var page []string
		for _, row := range rows {
			if h := rowHeight(row); h > band[0] && h < band[1] {
				page = append(page, row)
			}
		}
		if len(page) <= maxRows {
			respond(w, `{"data":{"getTransactions":[`+strings.Join(page, ",")+`]}}`)
			return
		}
		respond(w, `{"data":{"getTransactions":[`+strings.Join(page[:maxRows], ",")+
			`]},"errors":[{"message":"max elements per query reached"}]}`)
	}
}

var rowHeightPattern = regexp.MustCompile(`"block_height":(\d+)`)

func rowHeight(row string) int {
	h, _ := strconv.Atoi(rowHeightPattern.FindStringSubmatch(row)[1])
	return h
}

// txRow is a row of an indexer's answer, on r/alice/blog.
func txRow(hash, kind string, height int) string {
	return fmt.Sprintf(`{"hash":%q,"block_height":%d,"messages":[{"value":{"__typename":%q,"package":{"path":"gno.land/r/alice/blog"}}}]}`, hash, height, kind)
}

// One session message every 10 blocks, which the band's cap fills up on, and
// one deploy below where the cap stops the third band.
func TestRecentWalksOnPastACappedBand(t *testing.T) {
	const tip, every, maxRows = 300000, 10, 10000
	for _, deploy := range []int{160005, 100005} {
		t.Run(strconv.Itoa(deploy), func(t *testing.T) {
			var rows []string
			for h := tip; h > 0; h -= every {
				if h < deploy && h+every > deploy {
					rows = append(rows, txRow("deploy", "MsgAddPackage", deploy))
				}
				rows = append(rows, txRow("session", "UnexpectedMessage", h))
			}
			c, _ := newTestClient(t, cappedIndexer(t, tip, maxRows, rows))

			txs, err := c.Deploys(context.Background(), "gno.land/r/alice/blog", 20)
			if len(txs) != 1 || txs[0].Hash != "deploy" {
				t.Fatalf("txs = %+v, err = %v, want the deploy at %d", txs, err, deploy)
			}
		})
	}
}

// An indexer may end a capped page inside a block: an older tx-indexer cuts
// at the cap, a newer one at a block boundary. Neither may cost a row or
// return one twice.
func TestRecentResumesACappedBandAtAWholeBlock(t *testing.T) {
	deploy := func(hash string, h int) string { return txRow(hash, "MsgAddPackage", h) }
	// The cap of 3 falls between b and c, both in block 500.
	rows := []string{deploy("z", 700), deploy("a", 600), deploy("b", 500), deploy("c", 500), deploy("d", 400)}
	c, _ := newTestClient(t, cappedIndexer(t, 1000, 3, rows))

	txs, err := c.Deploys(context.Background(), "gno.land/r/alice/blog", 20)
	if err != nil {
		t.Fatalf("Deploys: %v", err)
	}
	var got []string
	for _, tx := range txs {
		got = append(got, tx.Hash)
	}
	if want := []string{"z", "a", "b", "c", "d"}; !slices.Equal(got, want) {
		t.Fatalf("rows = %v, want %v", got, want)
	}
}

// A walk that ends on a capped page has no band left to read its last block
// again, and rows in hand must not go with it.
func TestRecentKeepsACappedPagesLastBlockWhenTheWalkEnds(t *testing.T) {
	const tip = 1500000
	// The fourth and last band is capped, and its page ends on the deploy.
	var rows []string
	for h := tip - 150000; h > tip-150040; h -= 10 {
		rows = append(rows, txRow("session", "UnexpectedMessage", h))
	}
	rows = append(rows, txRow("deploy", "MsgAddPackage", tip-150040))
	rows = append(rows, txRow("session", "UnexpectedMessage", tip-150050))
	c, _ := newTestClient(t, cappedIndexer(t, tip, 5, rows))

	txs, err := c.Deploys(context.Background(), "gno.land/r/alice/blog", 20)
	if !errors.Is(err, ErrPartial) {
		t.Fatalf("err = %v, want ErrPartial", err)
	}
	if len(txs) != 1 || txs[0].Hash != "deploy" {
		t.Fatalf("txs = %+v, want the deploy the page held", txs)
	}
}

// One block filling the cap cannot be paged by height: the walk stops there
// and says so, rather than asking for the same block again.
func TestRecentStopsOnABlockFillingTheCap(t *testing.T) {
	deploy := func(hash string) string { return txRow(hash, "MsgAddPackage", 900) }
	rows := []string{deploy("a"), deploy("b"), deploy("c")}
	c, _ := newTestClient(t, cappedIndexer(t, 1000, 2, rows))

	txs, err := c.Deploys(context.Background(), "gno.land/r/alice/blog", 20)
	if !errors.Is(err, ErrPartial) {
		t.Fatalf("err = %v, want ErrPartial", err)
	}
	if len(txs) != 2 {
		t.Fatalf("txs = %+v, want the two rows the cap kept", txs)
	}
}

// Each clause of a where must name its message's typeUrl, which is what drops
// a message the indexer could not decode, and its match must keep the rows
// that clause asks for: a clause match does not know drops every row silently.
func TestEveryWhereClauseSurvivesItsMatch(t *testing.T) {
	const pkg, addr = "gno.land/r/demo/boards", "g1test"
	calls := map[string]func(*Client) ([]Tx, error){
		"RecentByPackage":  func(c *Client) ([]Tx, error) { return c.RecentByPackage(context.Background(), pkg, 20) },
		"RecentByAddress":  func(c *Client) ([]Tx, error) { return c.RecentByAddress(context.Background(), addr, 20) },
		"Deploys":          func(c *Client) ([]Tx, error) { return c.Deploys(context.Background(), pkg, 20) },
		"SourceContains":   func(c *Client) ([]Tx, error) { return c.SourceContains(context.Background(), "x", "", 20) },
		"DeploysImporting": func(c *Client) ([]Tx, error) { return c.DeploysImporting(context.Background(), pkg, 20) },
	}
	clause := regexp.MustCompile(`typeUrl: \{ eq: "\w+" \} value: \{ (\w+):`)
	for name, call := range calls {
		t.Run(name, func(t *testing.T) {
			var kinds []string
			c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				var req gqlRequest
				_ = json.Unmarshal(body, &req)
				if strings.Contains(req.Query, "latestBlockHeight") {
					respond(w, `{"data":{"latestBlockHeight":100}}`)
					return
				}
				if n, want := len(clause.FindAllString(req.Query, -1)), strings.Count(req.Query, "value: {"); n != want {
					t.Errorf("%d of %d clauses name their typeUrl:\n%s", n, want, req.Query)
				}
				var rows []string
				for _, m := range clause.FindAllStringSubmatch(req.Query, -1) {
					kinds = append(kinds, m[1])
					rows = append(rows, fmt.Sprintf(`{"hash":%q,"messages":[{"value":{"__typename":%q,"pkg_path":%q,"caller":%q,"creator":%q,"from_address":%q,"package":{"path":%q}}}]}`,
						m[1], m[1], pkg, addr, addr, addr, pkg))
				}
				respond(w, `{"data":{"getTransactions":[`+strings.Join(rows, ",")+`]}}`)
			})

			txs, err := call(c)
			if err != nil {
				t.Fatalf("%s: %v", name, err)
			}
			if len(kinds) == 0 || len(txs) != len(kinds) {
				t.Fatalf("kept %d of the %d rows asked for (%v)", len(txs), len(kinds), kinds)
			}
		})
	}
}

// Each kind of message names the address in its own field.
func TestRecentByAddressKeepsRowsNamingTheAddress(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req gqlRequest
		_ = json.Unmarshal(body, &req)
		if strings.Contains(req.Query, "latestBlockHeight") {
			respond(w, `{"data":{"latestBlockHeight":10}}`)
			return
		}
		respond(w, `{"data":{"getTransactions":[
			{"hash":"call","messages":[{"value":{"__typename":"MsgCall","caller":"g1a"}}]},
			{"hash":"run","messages":[{"value":{"__typename":"MsgRun","caller":"g1a"}}]},
			{"hash":"deploy","messages":[{"value":{"__typename":"MsgAddPackage","creator":"g1a"}}]},
			{"hash":"received","messages":[{"value":{"__typename":"BankMsgSend","from_address":"g1b","to_address":"g1a"}}]},
			{"hash":"someone-else","messages":[{"value":{"__typename":"MsgCall","caller":"g1b"}}]},
			{"hash":"session","messages":[{"value":{"__typename":"UnexpectedMessage"}}]}
		]}}`)
	})

	txs, err := c.RecentByAddress(context.Background(), "g1a", 10)
	if err != nil {
		t.Fatalf("RecentByAddress: %v", err)
	}
	var got []string
	for _, tx := range txs {
		got = append(got, tx.Hash)
	}
	if want := "call run deploy received"; strings.Join(got, " ") != want {
		t.Fatalf("hashes = %q, want %q", strings.Join(got, " "), want)
	}
}

// likeValues pulls every `like: "..."` literal back out of a query, decoded
// the way the indexer reads it.
func likeValues(t *testing.T, q string) []string {
	t.Helper()
	var out []string
	for _, m := range regexp.MustCompile(`like: ("(?:[^"\\]|\\.)*")`).FindAllStringSubmatch(q, -1) {
		var v string
		if err := json.Unmarshal([]byte(m[1]), &v); err != nil {
			t.Fatalf("like literal %s: %v", m[1], err)
		}
		out = append(out, v)
	}
	return out
}

func captureTxQuery(t *testing.T, run func(c *Client)) string {
	t.Helper()
	var txQuery string
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req gqlRequest
		_ = json.Unmarshal(body, &req)
		if strings.Contains(req.Query, "latestBlockHeight") {
			respond(w, `{"data":{"latestBlockHeight":10}}`)
			return
		}
		txQuery = req.Query
		respond(w, `{"data":{"getTransactions":[]}}`)
	})
	run(c)
	return txQuery
}

// tx-indexer evaluates `like` with regexp.MatchString, and a pattern that
// does not compile matches nothing. The term must reach it quoted.
func TestSourceContainsQuotesTheTerm(t *testing.T) {
	for _, term := range []string{"Render(", "[]byte", "*avl.Tree", "a.b", "zzzz|Render"} {
		t.Run(term, func(t *testing.T) {
			q := captureTxQuery(t, func(c *Client) {
				_, _ = c.SourceContains(context.Background(), term, "", 5)
			})
			likes := likeValues(t, q)
			if len(likes) != 1 {
				t.Fatalf("like filters = %q, want one", likes)
			}
			if likes[0] != regexp.QuoteMeta(term) {
				t.Fatalf("like = %q, want %q", likes[0], regexp.QuoteMeta(term))
			}
			re := regexp.MustCompile(likes[0])
			if !re.MatchString("x " + term + " y") {
				t.Errorf("%q does not match a body containing %q", likes[0], term)
			}
		})
	}

	// The metacharacters must not keep their meaning.
	q := captureTxQuery(t, func(c *Client) {
		_, _ = c.SourceContains(context.Background(), "a.b", "", 5)
	})
	if regexp.MustCompile(likeValues(t, q)[0]).MatchString("aXb") {
		t.Error(`"a.b" matches "aXb"`)
	}
}

// The author narrows the indexer's own filter, by creator address or by
// namespace, so older packages by that author are not crowded out of the
// capped answer by newer ones from others.
func TestSourceContainsFiltersAuthorOnTheIndexer(t *testing.T) {
	q := captureTxQuery(t, func(c *Client) {
		_, _ = c.SourceContains(context.Background(), "avl.Tree", "bob", 5)
	})
	likes := likeValues(t, q)
	if len(likes) != 3 {
		t.Fatalf("like filters = %q, want body, creator and path", likes)
	}
	if !strings.Contains(q, "_or:") {
		t.Fatalf("author is not an alternative of creator and path:\n%s", q)
	}
	creator, path := regexp.MustCompile(likes[1]), regexp.MustCompile(likes[2])

	for _, tt := range []struct {
		re   *regexp.Regexp
		in   string
		want bool
	}{
		{creator, "bob", true},
		{creator, "BOB", true},
		{creator, "bobby", false},
		{path, "gno.land/r/bob/counter", true},
		{path, "gno.land/p/bob", true},
		{path, "gno.land/r/bobby/counter", false},
		{path, "gno.land/r/alice/bob", false},
	} {
		if got := tt.re.MatchString(tt.in); got != tt.want {
			t.Errorf("%q.MatchString(%q) = %v, want %v", tt.re, tt.in, got, tt.want)
		}
	}

	// A metacharacter in the author is literal too.
	q = captureTxQuery(t, func(c *Client) {
		_, _ = c.SourceContains(context.Background(), "avl.Tree", "b.b", 5)
	})
	if regexp.MustCompile(likeValues(t, q)[2]).MatchString("gno.land/r/bxb/x") {
		t.Error(`author "b.b" matches namespace "bxb"`)
	}
}

// The endpoint is shown to every anonymous reader, and an operator's only
// place for a basic-auth or query-string credential is the URL itself.
func TestURLIsRedacted(t *testing.T) {
	for raw, want := range map[string]string{
		"https://ops:s3cret@indexer.example/graphql/query":      "https://indexer.example",
		"https://indexer.example/graphql/query?apikey=s3cret":   "https://indexer.example",
		"http://127.0.0.1:8546/graphql/query":                   "http://127.0.0.1:8546",
		"https://indexer.example/s3cret-path-key/graphql/query": "https://indexer.example",
		"not a url": "(indexer)",
	} {
		if got := New(raw, "").URL(); got != want {
			t.Errorf("URL() for %q = %q, want %q", raw, got, want)
		}
	}
}

// Transport errors are logged; net/http quotes the request URL in them and
// masks a password but not a query-string key.
func TestTransportErrorIsRedacted(t *testing.T) {
	c := New("http://ops:s3cret@127.0.0.1:1/graphql/query?apikey=hunter2", "")
	_, err := c.LatestBlockHeight(context.Background())
	if err == nil {
		t.Fatal("want a connection error")
	}
	for _, secret := range []string{"s3cret", "hunter2", "ops"} {
		if strings.Contains(err.Error(), secret) {
			t.Errorf("error %q carries %q", err, secret)
		}
	}
}

func TestValidateURL(t *testing.T) {
	for raw, ok := range map[string]bool{
		"https://indexer.example/graphql/query": true,
		"http://127.0.0.1:8546/graphql/query":   true,
		"localhost:8546/graphql":                false,
		"indexer.example/graphql":               false,
		"ftp://indexer.example/graphql":         false,
		"http:///graphql":                       false,
	} {
		if err := ValidateURL(raw); (err == nil) != ok {
			t.Errorf("ValidateURL(%q) = %v, want ok=%v", raw, err, ok)
		}
	}
}

// ValidateURL's error is printed at startup, so it quotes nothing of the URL.
func TestValidateURLErrorIsRedacted(t *testing.T) {
	for _, raw := range []string{
		"https://ops:s3cret@indexer example/graphql/query?apikey=hunter2",
		"https://ops:s3cret/x@indexer.example/graphql/query",
		// Scheme-less: the user part parses as the scheme.
		"ops:s3cret@indexer.example/graphql/query",
		"hunter2:@indexer.example/graphql/query",
	} {
		err := ValidateURL(raw)
		if err == nil {
			t.Fatalf("ValidateURL(%q): want an error", raw)
		}
		for _, secret := range []string{"ops", "s3cret", "hunter2"} {
			if strings.Contains(err.Error(), secret) {
				t.Errorf("error %q carries %q", err, secret)
			}
		}
	}
}

// A caller's deadline is the caller's budget: the omnibar gives up at 3s,
// below the client's 4s, and three readers typing at once must not close a
// healthy indexer to everyone for the cooldown. Scaled down: the caller gives
// up at 50ms, the indexer answers in 300ms.
func TestCallerDeadlineDoesNotOpenBreaker(t *testing.T) {
	c, calls := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(300 * time.Millisecond):
		}
		respond(w, `{"data":{"latestBlockHeight":5}}`)
	})

	var out struct {
		H int `json:"latestBlockHeight"`
	}
	for range breakerThreshold + 1 {
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		err := c.Query(ctx, `{ latestBlockHeight }`, &out)
		cancel()
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("err = %v, want the caller's deadline", err)
		}
	}

	if err := c.Query(context.Background(), `{ latestBlockHeight }`, &out); err != nil {
		t.Fatalf("next caller: %v, want an answer", err)
	}
	if calls.Load() != breakerThreshold+2 {
		t.Fatalf("calls = %d, want every query sent", calls.Load())
	}
}

// GraphQL answers `data`, `errors` or both. A body with neither is what an
// -indexer-url pointing at the wrong endpoint returns; it must not read as
// "not found" or a tip of 0.
func TestEnvelopeWithoutDataIsAnError(t *testing.T) {
	for _, body := range []string{
		`{"jsonrpc":"2.0","id":null,"error":{"code":-32600,"message":"invalid request"}}`,
		`{"data":null}`,
		`{}`,
	} {
		c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			respond(w, body)
		})
		if h, err := c.LatestBlockHeight(context.Background()); err == nil {
			t.Errorf("%s: LatestBlockHeight = %d, nil", body, h)
		}
		if _, err := c.TxByHash(context.Background(), "deadbeef"); err == nil || errors.Is(err, ErrNotFound) {
			t.Errorf("%s: TxByHash err = %v, want a failure that is not ErrNotFound", body, err)
		}
	}
}

// tx-indexer evaluates `like` with regexp.MatchString. The importers pattern
// must match the path as an import quotes it, not inside a longer path.
func TestDeploysImportingMatchesTheQuotedPathOnly(t *testing.T) {
	q := captureTxQuery(t, func(c *Client) {
		_, _ = c.DeploysImporting(context.Background(), "gno.land/r/demo/foo", 20)
	})
	likes := likeValues(t, q)
	if len(likes) != 1 {
		t.Fatalf("like filters = %q", likes)
	}
	like := regexp.MustCompile(likes[0])
	for body, want := range map[string]bool{
		`import "gno.land/r/demo/foo"`:       true,
		"import foo `gno.land/r/demo/foo`":   true,
		`import "gno.land/r/demo/foobar"`:    false,
		`module = "gno.land/r/demo/foo/sub"`: false,
		`import "gno.land/r/demo/foo/sub"`:   false,
		`import "gnoXland/r/demo/foo"`:       false,
	} {
		if got := like.MatchString(body); got != want {
			t.Errorf("like %q on %q = %v, want %v", likes[0], body, got, want)
		}
	}
}

// A walk that runs out of steps before genesis says so: a deploy below the
// lowest band must not read as "never deployed".
func TestRecentReportsAWalkStoppedShortOfGenesis(t *testing.T) {
	bands := 0
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req gqlRequest
		_ = json.Unmarshal(body, &req)
		if strings.Contains(req.Query, "latestBlockHeight") {
			respond(w, `{"data":{"latestBlockHeight":2000000}}`)
			return
		}
		bands++
		respond(w, `{"data":{"getTransactions":[]}}`)
	})

	txs, err := c.Deploys(context.Background(), "gno.land/r/demo/boards", 20)
	if !errors.Is(err, ErrPartial) {
		t.Fatalf("Deploys: rows=%d err=%v, want ErrPartial", len(txs), err)
	}
	if bands != maxWindowSteps {
		t.Fatalf("bands = %d, want %d", bands, maxWindowSteps)
	}

	// Reaching genesis is the whole answer, however few rows it holds.
	c, _ = newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if strings.Contains(string(body), "latestBlockHeight") {
			respond(w, `{"data":{"latestBlockHeight":1500}}`)
			return
		}
		respond(w, `{"data":{"getTransactions":[]}}`)
	})
	if _, err := c.Deploys(context.Background(), "gno.land/r/demo/boards", 20); err != nil {
		t.Fatalf("Deploys down to genesis: %v, want no error", err)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// The tip fetch carries no deadline of its own. One armed next to the
// client's equal timeout always fires first, reads as the caller giving up,
// and a hung indexer never opens the breaker.
func TestTipFetchEndsOnTheClientTimeout(t *testing.T) {
	c := New("http://indexer.invalid/graphql/query", "")
	c.http.Timeout = time.Hour
	var left time.Duration
	c.http.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		d, _ := r.Context().Deadline()
		left = time.Until(d)
		return nil, errors.New("down")
	})

	// A leader with a short deadline must not hand it to the shared fetch.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, _ = c.LatestBlockHeight(ctx)
	if left < time.Minute {
		t.Fatalf("tip fetch deadline in %v, want the client's hour", left)
	}
}

// A band slower than the client's timeout on a healthy indexer is the band's
// width, not the indexer's health: it must not close every other lookup for
// the cooldown. The tip fetch, which every windowed search makes first, is
// the health signal. Scaled down: the client gives up at 100ms.
func TestSlowBandDoesNotOpenBreaker(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		switch {
		case strings.Contains(string(body), "latestBlockHeight"):
			respond(w, `{"data":{"latestBlockHeight":100000}}`)
		case strings.Contains(string(body), "block_height: {"):
			select {
			case <-r.Context().Done():
			case <-time.After(time.Second):
			}
		default:
			respond(w, `{"data":{"getTransactions":[{"hash":"abc"}]}}`)
		}
	})
	c.http.Timeout = 100 * time.Millisecond

	for range breakerThreshold {
		if _, err := c.SourceContains(context.Background(), "avl.Tree", "", 5); err == nil {
			t.Fatal("SourceContains: want the band's timeout")
		}
	}
	if _, err := c.TxByHash(context.Background(), "abc"); err != nil {
		t.Fatalf("TxByHash after slow bands: %v, want an answer", err)
	}
}

// gqlString is the only guard on a raw `tx:` or `account:` term: every `eq:`
// value must reach the indexer as one string literal holding the input.
func TestEqValuesAreEscaped(t *testing.T) {
	const hostile = `x" } } }) { hash } evil: getBlocks { hash } # \`
	eqLiteral := regexp.MustCompile(`eq: ("(?:[^"\\]|\\.)*")`)
	typeURLClause := regexp.MustCompile(`typeUrl: \{ eq: "\w+" \}`)
	for name, run := range map[string]func(c *Client){
		"TxByHash":        func(c *Client) { _, _ = c.TxByHash(context.Background(), hostile) },
		"RecentByAddress": func(c *Client) { _, _ = c.RecentByAddress(context.Background(), hostile, 5) },
		"RecentByPackage": func(c *Client) { _, _ = c.RecentByPackage(context.Background(), hostile, 5) },
		"Deploys":         func(c *Client) { _, _ = c.Deploys(context.Background(), hostile, 5) },
	} {
		t.Run(name, func(t *testing.T) {
			// A typeUrl is a constant of this package, not an input.
			q := typeURLClause.ReplaceAllString(captureTxQuery(t, run), "")
			eqs := eqLiteral.FindAllStringSubmatch(q, -1)
			if len(eqs) == 0 || strings.Count(q, "eq:") != len(eqs) {
				t.Fatalf("eq: values are not all string literals:\n%s", q)
			}
			for _, m := range eqs {
				var v string
				if err := json.Unmarshal([]byte(m[1]), &v); err != nil || v != hostile {
					t.Fatalf("eq: %s decodes to %q (%v), want the input", m[1], v, err)
				}
			}
		})
	}
}

// A redirect would aim gnoweb at a host nobody configured.
func TestRedirectIsNotFollowed(t *testing.T) {
	var followed atomic.Int32
	elsewhere := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		followed.Add(1)
		respond(w, `{"data":{"latestBlockHeight":1}}`)
	}))
	t.Cleanup(elsewhere.Close)
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, elsewhere.URL, http.StatusFound)
	})

	if _, err := c.LatestBlockHeight(context.Background()); err == nil {
		t.Fatal("LatestBlockHeight: want the redirect refused")
	}
	if followed.Load() != 0 {
		t.Fatal("the redirect was followed")
	}
}

// countingBody yields four times the cap of 'a' and counts what was read.
type countingBody struct{ n int }

func (b *countingBody) Read(p []byte) (int, error) {
	if b.n >= 4*maxResponseSize {
		return 0, io.EOF
	}
	p = p[:min(len(p), 4*maxResponseSize-b.n)]
	for i := range p {
		p[i] = 'a'
	}
	b.n += len(p)
	return len(p), nil
}
func (*countingBody) Close() error { return nil }

// The cap bounds the read itself, not just the decode: an oversized body is
// read no further than one byte past it.
func TestResponseReadIsBounded(t *testing.T) {
	body := &countingBody{}
	c := New("http://indexer.invalid/graphql/query", "")
	c.http.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Body: body, Header: http.Header{}}, nil
	})

	var out struct{}
	if err := c.Query(context.Background(), `{ latestBlockHeight }`, &out); !errors.Is(err, ErrResponseTooLarge) {
		t.Fatalf("error = %v, want ErrResponseTooLarge", err)
	}
	if body.n > maxResponseSize+1 {
		t.Fatalf("read %d bytes, want at most %d", body.n, maxResponseSize+1)
	}
}

// At most maxConcurrent requests are in flight; the next waits for a slot.
func TestInFlightRequestsAreCapped(t *testing.T) {
	release := make(chan struct{})
	var inFlight atomic.Int32
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		inFlight.Add(1)
		<-release
		respond(w, `{"data":{"latestBlockHeight":1}}`)
	})

	var wg sync.WaitGroup
	for range maxConcurrent + 1 {
		wg.Go(func() {
			var out struct{}
			_ = c.Query(context.Background(), `{ latestBlockHeight }`, &out)
		})
	}
	time.Sleep(200 * time.Millisecond)
	got := inFlight.Load()
	close(release)
	wg.Wait()
	if got != maxConcurrent {
		t.Fatalf("in flight = %d, want %d", got, maxConcurrent)
	}
}
