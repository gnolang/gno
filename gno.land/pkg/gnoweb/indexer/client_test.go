package indexer

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

// newTestClient returns a Client pointed at a stub indexer, plus a pointer to
// the request counter so a test can assert that the breaker really skips the
// network rather than just swallowing the error.
func newTestClient(t *testing.T, handler http.HandlerFunc) (*Client, *int) {
	t.Helper()
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
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
	if *calls != 1 {
		t.Fatalf("requests = %d, want 1 (height must be cached within heightTTL)", *calls)
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
	if *calls != breakerThreshold {
		t.Fatalf("requests before breaker = %d, want %d", *calls, breakerThreshold)
	}

	err := c.Query(context.Background(), `{ latestBlockHeight }`, &out)
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("error = %v, want ErrUnavailable", err)
	}
	if *calls != breakerThreshold {
		t.Fatalf("requests after breaker = %d, want it to stay at %d", *calls, breakerThreshold)
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
	if *calls != breakerThreshold+2 {
		t.Fatalf("requests = %d, want all %d to reach the network", *calls, breakerThreshold+2)
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

func TestRecentWidensWindowUntilEnough(t *testing.T) {
	var queries []string
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req gqlRequest
		_ = json.Unmarshal(body, &req)
		queries = append(queries, req.Query)

		if strings.Contains(req.Query, "latestBlockHeight") {
			respond(w, `{"data":{"latestBlockHeight":100000}}`)
			return
		}
		// Only the widest window holds a row, so the loop has to widen.
		if strings.Contains(req.Query, "gt: 98000") {
			respond(w, `{"data":{"getTransactions":[]}}`)
			return
		}
		respond(w, `{"data":{"getTransactions":[{"hash":"abc"}]}}`)
	})

	txs, err := c.RecentByPackage(context.Background(), "gno.land/r/demo/boards", 1)
	if err != nil {
		t.Fatalf("RecentByPackage: %v", err)
	}
	if len(txs) != 1 {
		t.Fatalf("txs = %d, want 1 after widening", len(txs))
	}
	if len(queries) != 3 {
		t.Fatalf("queries = %d (%v), want 3: height, narrow window, widened window", len(queries), queries)
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
	if *calls != 1 {
		t.Fatalf("requests = %d, want 1 for %d concurrent callers", *calls, callers)
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
