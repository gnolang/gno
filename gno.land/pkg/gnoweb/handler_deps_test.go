package gnoweb_test

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/indexer"
)

// stubIndexer is the *indexer.Client surface both indexer-backed features
// read. Only the importer search answers; the rest is never reached here.
type stubIndexer struct{ deploys []indexer.Tx }

var errUnused = errors.New("not used by these tests")

func (s stubIndexer) LatestBlockHeight(context.Context) (int, error) { return 77, nil }
func (s stubIndexer) URL() string                                    { return "https://indexer.test/graphql/query" }
func (s stubIndexer) DeploysQuoting(_ context.Context, _ string, lower, _ int) ([]indexer.Tx, error) {
	if lower >= 0 {
		return nil, nil // the stub's deploys sit at height 0, in the bottom band
	}
	return s.deploys, nil
}

func (stubIndexer) TxByHash(context.Context, string) (*indexer.Tx, error) { return nil, errUnused }
func (stubIndexer) RecentByPackage(context.Context, string, int) ([]indexer.Tx, error) {
	return nil, errUnused
}
func (stubIndexer) RecentByAddress(context.Context, string, int) ([]indexer.Tx, error) {
	return nil, errUnused
}
func (stubIndexer) Deploys(context.Context, string, int) ([]indexer.Tx, error) {
	return nil, errUnused
}
func (stubIndexer) SourceContains(context.Context, string, string, int) ([]indexer.Tx, error) {
	return nil, errUnused
}
func (stubIndexer) Block(context.Context, int) (*indexer.Block, error) { return nil, errUnused }
func (stubIndexer) CallsBetween(context.Context, int, int) ([]indexer.Tx, error) {
	return nil, errUnused
}

func deployOf(path string) indexer.Tx {
	tx := indexer.Tx{Messages: make([]indexer.Message, 1)}
	tx.Messages[0].Value.Type = "MsgAddPackage"
	tx.Messages[0].Value.Package = &struct {
		Path string `json:"path"`
	}{Path: path}
	return tx
}

func newDepsHandler(t *testing.T, idx *stubIndexer) http.Handler {
	t.Helper()
	client := gnoweb.NewMockClient(
		&gnoweb.MockPackage{Path: "/p/demo/lib", Files: map[string]string{"lib.gno": "package lib"}, Imports: []string{"strings"}},
		&gnoweb.MockPackage{Path: "/r/demo/user", Files: map[string]string{"user.gno": "package user"}, Imports: []string{"/p/demo/lib"}},
		&gnoweb.MockPackage{Path: "/r/demo/mention", Files: map[string]string{"m.gno": "package m"}, Imports: []string{"strings"}},
	)
	cfg := newTestHandlerConfig(t, client)
	if idx != nil {
		cfg.Indexer = idx
	}
	logger := slog.New(slog.NewTextHandler(&testingLogger{t}, &slog.HandlerOptions{}))
	h, err := gnoweb.NewHTTPHandler(logger, cfg)
	if err != nil {
		t.Fatalf("NewHTTPHandler: %v", err)
	}
	return h
}

// With no indexer nothing anywhere mentions importers: not on the overview,
// not on the dependencies page, which still draws the imports.
func TestHTTPHandler_DepsWithoutIndexer(t *testing.T) {
	t.Parallel()

	h := newDepsHandler(t, nil)
	overview := serve(t, h, "/r/demo/user$source").Body.String()
	if !strings.Contains(overview, "b-depgraph") || !strings.Contains(overview, `href="/p/demo/lib"`) {
		t.Error("the overview must draw the imports as a graph")
	}
	if strings.Contains(overview, "$deps") || strings.Contains(overview, "Imported by") {
		t.Error("without an indexer the overview must not offer importers")
	}

	rr := serve(t, h, "/p/demo/lib$deps")
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	if body := rr.Body.String(); strings.Contains(body, "Imported by") || strings.Contains(body, "b-tag--indexer") {
		t.Error("without an indexer the dependencies page must not mention importers")
	}
}

// The indexer proposes, the chain decides: a deploy that only mentions the
// path is dropped because its qdoc imports do not name it.
func TestHTTPHandler_DepsListsCheckedImporters(t *testing.T) {
	t.Parallel()

	idx := &stubIndexer{deploys: []indexer.Tx{deployOf("/r/demo/user"), deployOf("/r/demo/mention"), deployOf("/r/demo/gone")}}
	h := newDepsHandler(t, idx)

	overview := serve(t, h, "/p/demo/lib$source").Body.String()
	if !strings.Contains(overview, `href="/p/demo/lib$deps"`) {
		t.Error("with an indexer the overview must link to the importers")
	}

	rr := serve(t, h, "/p/demo/lib$deps")
	body := rr.Body.String()
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	if !strings.Contains(body, `<span class="b-tag--tertiary">1</span>`) || !strings.Contains(body, `href="/r/demo/user"`) {
		t.Error("the checked importer is missing")
	}
	if strings.Contains(body, `href="/r/demo/mention"`) || strings.Contains(body, `href="/r/demo/gone"`) {
		t.Error("a candidate the chain does not confirm must not be listed")
	}
	if !strings.Contains(body, "last indexed block 77") {
		t.Error("the page must state the indexer's freshness")
	}
	if !strings.Contains(body, `content="noindex`) {
		t.Error("a page costing an indexer scan must not be offered to crawlers")
	}
}

func TestHTTPHandler_DepsOfAMissingPackage(t *testing.T) {
	t.Parallel()

	if got := serve(t, newDepsHandler(t, nil), "/p/demo/nothing$deps").Code; got != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", got)
	}
}
