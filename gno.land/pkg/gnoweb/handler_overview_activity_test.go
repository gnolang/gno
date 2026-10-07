package gnoweb_test

import (
	"log/slog"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/indexer"
)

func overviewHandler(t *testing.T, client *gnoweb.MockClient, idx *stubIndexer) http.Handler {
	t.Helper()
	cfg := newTestHandlerConfig(t, client)
	if idx != nil {
		cfg.Indexer = idx
	}
	h, err := gnoweb.NewHTTPHandler(slog.New(slog.NewTextHandler(&testingLogger{t}, nil)), cfg)
	if err != nil {
		t.Fatalf("NewHTTPHandler: %v", err)
	}
	return h
}

func mockPkg(path string, storage *gnoweb.RealmStorage) *gnoweb.MockPackage {
	return &gnoweb.MockPackage{Path: path, Files: map[string]string{"a.gno": "package a"}, Storage: storage}
}

// A realm's sidebar states what it keeps on chain and the GNOT locked for it,
// read from the node: no indexer involved.
func TestHTTPHandler_OverviewShowsStorage(t *testing.T) {
	t.Parallel()

	client := gnoweb.NewMockClient(mockPkg("/r/demo/wugnot", &gnoweb.RealmStorage{Bytes: 1292654, Deposit: 129265400}))
	body := serve(t, overviewHandler(t, client, nil), "/r/demo/wugnot$source").Body.String()
	for _, want := range []string{">Storage</dt>", ">1.29 MB<", ">Storage deposit</dt>", ">129.27 GNOT<"} {
		if !strings.Contains(body, want) {
			t.Errorf("overview lacks %q", want)
		}
	}
}

// A pure package keeps no state: the node is not asked, and nothing is shown.
// A realm the node cannot answer for shows no rows rather than zeros.
func TestHTTPHandler_OverviewStorageOnlyForRealms(t *testing.T) {
	t.Parallel()

	var asked atomic.Int32
	client := gnoweb.NewMockClient(mockPkg("/p/demo/lib", &gnoweb.RealmStorage{Bytes: 1, Deposit: 1}), mockPkg("/r/demo/quiet", nil))
	client.OnStorage = func(string) { asked.Add(1) }
	h := overviewHandler(t, client, nil)
	if body := serve(t, h, "/p/demo/lib$source").Body.String(); strings.Contains(body, "Storage deposit") || asked.Load() != 0 {
		t.Errorf("a pure package must neither query nor show storage (queries: %d)", asked.Load())
	}
	rr := serve(t, h, "/r/demo/quiet$source")
	if rr.Code != http.StatusOK || strings.Contains(rr.Body.String(), "Storage deposit") {
		t.Error("a realm whose storage the node cannot give must render without the rows")
	}
}

func callMsg(pkg, fn, caller string) indexer.Message {
	var m indexer.Message
	m.Value.Type, m.Value.PkgPath, m.Value.Func, m.Value.Caller = "MsgCall", pkg, fn, caller
	return m
}

// With an indexer the overview lists the realm's last calls, taken from the
// 7-day activity scan the map already makes: no query of its own. Each row
// says the function, the full caller, the transaction's gas and how many
// calls that transaction held, and whether it failed.
func TestHTTPHandler_OverviewRecentCalls(t *testing.T) {
	t.Parallel()

	const pkg = "/r/demo/wugnot" // the test config has no domain
	var run indexer.Message
	run.Value.Type = "MsgRun"
	const caller = "g13959x7zm49jaeyrfeltjkwz8adu8p0r0uhffag"
	idx := &stubIndexer{calls: []indexer.Tx{
		{Height: 70, Success: true, GasUsed: 154_000_000, Messages: []indexer.Message{callMsg(pkg, "Approve", caller), callMsg(pkg, "Deposit", caller), callMsg("/r/other/x", "Swap", caller)}},
		{Height: 60, Success: false, GasUsed: 120_000, Messages: []indexer.Message{callMsg(pkg, "Withdraw", "g1wzlp8l9quwf8cfa357amqlw83gssnzuwh4zucu"), run}},
		{Height: 55, Success: true, GasUsed: 1_000, Messages: []indexer.Message{callMsg(pkg, "Transfer", caller)}},
		{Height: 50, Success: true, GasUsed: 9_000_000, Messages: []indexer.Message{callMsg("/r/other/x", "Swap", "g1zzz")}},
	}}
	client := gnoweb.NewMockClient(mockPkg(pkg, nil), mockPkg("/r/other/x", nil))
	body := serveWarm(t, overviewHandler(t, client, idx), pkg+"$source", "b-calls__row")
	for _, want := range []string{
		`id="calls"`, `href="#calls"`, ">Recent calls<",
		">Approve, Deposit<", caller, "154 M gas", "tx of 3 calls",
		">Withdraw<", `b-tag--failed`, "120 k gas",
		`<time datetime=`, ">block 70<", "b-tag--indexer", "last indexed block",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("overview lacks %q", want)
		}
	}
	if n := strings.Count(body, `class="b-calls__row"`); n != 3 || strings.Contains(body, ">Swap<") {
		t.Errorf("rows = %d, want one row per transaction into this realm", n)
	}
	if strings.Contains(body, "tx of 2 calls") {
		t.Error("a call batched with a run is not a tx of 2 calls")
	}
}

// Recent calls are only for a realm at the latest height: a pure package
// cannot be called, and a page pinned to a past height must not show calls
// made since. A realm with no call in the window says so.
func TestHTTPHandler_OverviewRecentCallsScope(t *testing.T) {
	t.Parallel()

	idx := &stubIndexer{calls: []indexer.Tx{}}
	client := gnoweb.NewMockClient(mockPkg("/r/demo/quiet", nil), mockPkg("/p/demo/lib", nil))
	h := overviewHandler(t, client, idx)

	if body := serveWarm(t, h, "/r/demo/quiet$source", "No calls in the last 7 days."); !strings.Contains(body, "No calls in the last 7 days.") {
		t.Error("a realm with no call in the window must say so")
	}
	if body := serve(t, h, "/p/demo/lib$source").Body.String(); strings.Contains(body, `id="calls"`) || strings.Contains(body, `href="#calls"`) {
		t.Error("a pure package must have no calls section")
	}
	if body := serve(t, h, "/r/demo/quiet$source&height=10").Body.String(); strings.Contains(body, `id="calls"`) {
		t.Error("a page pinned to a past height must have no calls section")
	}
	if body := serve(t, overviewHandler(t, client, nil), "/r/demo/quiet$source").Body.String(); strings.Contains(body, `id="calls"`) {
		t.Error("without an indexer the overview must not have a calls section")
	}
}
