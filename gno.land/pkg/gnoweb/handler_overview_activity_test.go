package gnoweb_test

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/indexer"
)

func overviewHandler(t *testing.T, storage *gnoweb.PackageStorage, idx *stubIndexer) http.Handler {
	t.Helper()
	pkg := &gnoweb.MockPackage{Path: "/r/demo/wugnot", Files: map[string]string{"wugnot.gno": "package wugnot"}, Storage: storage}
	cfg := newTestHandlerConfig(t, gnoweb.NewMockClient(pkg))
	if idx != nil {
		cfg.Indexer = idx
	}
	h, err := gnoweb.NewHTTPHandler(slog.New(slog.NewTextHandler(&testingLogger{t}, nil)), cfg)
	if err != nil {
		t.Fatalf("NewHTTPHandler: %v", err)
	}
	return h
}

// The sidebar states what the realm keeps on chain and the GNOT locked for it,
// read from the node: no indexer involved.
func TestHTTPHandler_OverviewShowsStorage(t *testing.T) {
	t.Parallel()

	body := serve(t, overviewHandler(t, &gnoweb.PackageStorage{Bytes: 1292654, Deposit: 129265400}, nil), "/r/demo/wugnot$source").Body.String()
	for _, want := range []string{">Storage</dt>", ">1.29 MB<", ">Deposit</dt>", ">129.27 GNOT<"} {
		if !strings.Contains(body, want) {
			t.Errorf("overview lacks %q", want)
		}
	}
}

// A node that cannot say leaves the rows out: a zero would read as an empty
// realm.
func TestHTTPHandler_OverviewWithoutStorage(t *testing.T) {
	t.Parallel()

	rr := serve(t, overviewHandler(t, nil, nil), "/r/demo/wugnot$source")
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	if strings.Contains(rr.Body.String(), ">Deposit</dt>") {
		t.Error("storage rows must be left out when the node does not answer")
	}
}

func callTx(height, gas int, ok bool, msgs ...indexer.Message) indexer.Tx {
	return indexer.Tx{Height: height, GasUsed: gas, Success: ok, Messages: msgs}
}

func msgCall(pkg, fn, caller string) indexer.Message {
	var m indexer.Message
	m.Value.Type, m.Value.PkgPath, m.Value.Func, m.Value.Caller = "MsgCall", pkg, fn, caller
	return m
}

// With an indexer the overview lists the realm's last calls: function, full
// caller, the transaction's gas, failures, and the calls batched with it.
// Deploys and rows the indexer could not decode are not calls.
func TestHTTPHandler_OverviewRecentCalls(t *testing.T) {
	t.Parallel()

	const pkg = "/r/demo/wugnot" // the test config has no domain
	deploy := indexer.Message{}
	deploy.Value.Type = "MsgAddPackage"
	undecoded := indexer.Message{}
	undecoded.Value.Type = "UnexpectedMessage"
	idx := &stubIndexer{
		recent: []indexer.Tx{
			callTx(120, 154_000_000, true, msgCall(pkg, "Approve", "g13959x7zm49jaeyrfeltjkwz8adu8p0r0uhffag"), msgCall("/r/other/x", "Swap", "g13959x7zm49jaeyrfeltjkwz8adu8p0r0uhffag")),
			callTx(110, 10_000_000, false, msgCall(pkg, "Withdraw", "g1wzlp8l9quwf8cfa357amqlw83gssnzuwh4zucu")),
			callTx(100, 5_000_000, true, deploy),
			callTx(90, 1, true, undecoded),
		},
		times: map[int]time.Time{120: time.Now().Add(-35 * time.Minute), 110: time.Now().Add(-3 * time.Hour)},
	}
	body := serve(t, overviewHandler(t, nil, idx), "/r/demo/wugnot$source").Body.String()
	for _, want := range []string{
		`id="calls"`, `href="#calls"`, "Recent calls",
		">Approve<", "g13959x7zm49jaeyrfeltjkwz8adu8p0r0uhffag", "154 M gas", "+1 other call", "35 min ago",
		`Withdraw <span class="b-tag b-calls__failed">failed</span>`, "3 h ago",
		"b-tag--indexer", "last indexed block",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("overview lacks %q", want)
		}
	}
	if strings.Contains(body, "MsgAddPackage") || strings.Count(body, `class="b-calls__row"`) != 2 {
		t.Errorf("only the two calls are rows, got %d", strings.Count(body, `class="b-calls__row"`))
	}
}

// Without an indexer nothing mentions calls; with one that fails, the section
// says so instead of showing an empty list that reads as "never called".
func TestHTTPHandler_OverviewRecentCallsStates(t *testing.T) {
	t.Parallel()

	if body := serve(t, overviewHandler(t, nil, nil), "/r/demo/wugnot$source").Body.String(); strings.Contains(body, `id="calls"`) {
		t.Error("without an indexer the overview must not have a calls section")
	}
	failing := &stubIndexer{recentErr: errors.New("indexer down")}
	body := serve(t, overviewHandler(t, nil, failing), "/r/demo/wugnot$source").Body.String()
	if !strings.Contains(body, "Recent calls unavailable") {
		t.Error("a failing indexer must be said, not shown as no calls")
	}
	empty := &stubIndexer{recent: []indexer.Tx{}}
	body = serve(t, overviewHandler(t, nil, empty), "/r/demo/wugnot$source").Body.String()
	if !strings.Contains(body, "No calls found") {
		t.Error("an empty answer must say no calls were found")
	}
}

var _ = context.Background
