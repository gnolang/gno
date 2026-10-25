package omnisearch

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/indexer"
	"github.com/gnolang/gno/gnovm/pkg/doc"
)

// Every cut a list goes through is said, so a partial list never reads as a
// complete one.

func TestCappedSelectorResultsAreFlagged(t *testing.T) {
	t.Parallel()

	jdoc := &doc.JSONDocumentation{}
	for i := range MaxResults + 5 {
		jdoc.Funcs = append(jdoc.Funcs, &doc.JSONFunc{
			Name: fmt.Sprintf("Vote%d", i), Crossing: true,
			Signature: fmt.Sprintf("func Vote%d(cur realm)", i), File: "boards.gno", Line: i + 1,
		})
	}
	h := newHandler(t, &mockClient{doc: jdoc}, nil)
	groups, _ := h.Search(context.Background(), mustQuery(t, h, "func:Vote", "/r/demo/boards"))
	if len(groups) != 1 || len(groups[0].Results) != MaxResults {
		t.Fatalf("groups = %+v, want one group of %d", groups, MaxResults)
	}
	if g := groups[0]; !g.Truncated || !strings.Contains(g.Notice, "first 20") {
		t.Fatalf("Truncated = %v, Notice = %q", g.Truncated, g.Notice)
	}
}

func TestDiscoveryCapSaysHowManyMatched(t *testing.T) {
	t.Parallel()

	dir := &mockDirectory{}
	for i := range 25 {
		dir.realms = append(dir.realms, fmt.Sprintf("gno.land/r/u%c/blog", 'a'+i))
	}
	h := newHandlerWithDir(t, newDiscoveryClient(), dir, nil)
	out := renderText(t, h.build(context.Background(), mustQuery(t, h, "blog", "")))
	for _, want := range []string{"Showing 10 of 25 matches.", "Users 10 Showing 10 of 25 matches."} {
		if !strings.Contains(out, want) {
			t.Errorf("page does not say %q:\n%s", want, out)
		}
	}
}

// discover drops empty groups; a capped listing with no match in its visible
// part still says the listing was capped.
func TestTruncatedListingWithNoMatchKeepsItsNotice(t *testing.T) {
	t.Parallel()

	dir := newDiscoveryDir()
	dir.truncated = true
	h := newHandlerWithDir(t, newDiscoveryClient(), dir, nil)
	out := renderText(t, h.build(context.Background(), mustQuery(t, h, "zzzz", "")))
	if !strings.Contains(out, "the node caps this listing") {
		t.Fatalf("no truncation notice:\n%s", out)
	}
}

// A node that answers none of the candidates is "Could not answer", never a
// realm lacking the text.
func TestRenderFailureIsNotNothingMatched(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct{ raw, scope string }{
		{"render:voting", "/r/alice/blog"},
		{"render:voting author:alice", ""},
	} {
		c := newDiscoveryClient() // no renders: every Realm call fails
		h := newHandlerWithDir(t, c, newDiscoveryDir(), nil)
		groups, _ := h.Search(context.Background(), mustQuery(t, h, tc.raw, tc.scope))
		if len(groups) != 1 || groups[0].Err == nil {
			t.Errorf("%s: groups = %+v, want a failed group", tc.raw, groups)
		}
	}
}

// Past maxRenderCandidates, or over a capped listing, the field was cut.
func TestRenderCapIsFlagged(t *testing.T) {
	t.Parallel()

	paths := make([]string, 0, maxRenderCandidates+1)
	renders := map[string]string{}
	for i := range maxRenderCandidates + 1 {
		p := fmt.Sprintf("gno.land/r/alice/app%c", 'a'+i)
		paths = append(paths, p)
		renders[strings.TrimPrefix(p, "gno.land")] = "nothing here"
	}
	h := newHandlerWithDir(t, &mockClient{renders: renders}, &mockDirectory{realms: paths}, nil)
	groups, _ := h.Search(context.Background(), mustQuery(t, h, "render:voting author:alice", ""))
	if len(groups) != 1 || !groups[0].Truncated || !strings.Contains(groups[0].Notice, "first 8") {
		t.Fatalf("groups = %+v, want the candidate cap flagged", groups)
	}

	dir := newDiscoveryDir()
	dir.truncated = true
	h = newHandlerWithDir(t, &mockClient{renders: map[string]string{"/r/alice/blog": "x"}}, dir, nil)
	groups, _ = h.Search(context.Background(), mustQuery(t, h, "render:voting author:alice", ""))
	if len(groups) != 1 || !groups[0].Truncated || !strings.Contains(groups[0].Notice, "caps the realm listing") {
		t.Fatalf("groups = %+v, want the listing cap flagged", groups)
	}
}

// Results come back in completion order; the page shows them sorted.
func TestRenderResultsAreSorted(t *testing.T) {
	t.Parallel()

	paths := []string{"gno.land/r/alice/c", "gno.land/r/alice/a", "gno.land/r/alice/b"}
	renders := map[string]string{"/r/alice/a": "voting", "/r/alice/b": "voting", "/r/alice/c": "voting"}
	for range 20 {
		h := newHandlerWithDir(t, &mockClient{renders: renders}, &mockDirectory{realms: paths}, nil)
		groups, _ := h.Search(context.Background(), mustQuery(t, h, "render:voting author:alice", ""))
		var got []string
		for _, r := range groups[0].Results {
			got = append(got, r.Title)
		}
		if strings.Join(got, ",") != "/r/alice/a,/r/alice/b,/r/alice/c" {
			t.Fatalf("order = %v", got)
		}
	}
}

func TestSnippetKeepsRunesWhole(t *testing.T) {
	t.Parallel()

	body := []byte(strings.Repeat("é", 100) + "voting" + strings.Repeat("é", 100))
	from := strings.Index(string(body), "voting")
	for shift := range 2 {
		s := snippetAround(body[shift:], from-shift, from-shift+len("voting"))
		if strings.ContainsRune(s, '\uFFFD') {
			t.Fatalf("shift %d: snippet splits a rune: %q", shift, s)
		}
	}
}

// An indexer walk that stopped short of genesis keeps its rows and says so.
func TestPartialIndexerWalkIsFlagged(t *testing.T) {
	t.Parallel()

	idx := &mockIndexer{err: indexer.ErrPartial}
	h := newHandler(t, &mockClient{}, idx)
	groups, _ := h.Search(context.Background(), mustQuery(t, h, "deploys", "/r/demo/boards"))
	if len(groups) != 1 || groups[0].Err != nil || !groups[0].Truncated || groups[0].Notice != recentNotice {
		t.Fatalf("groups = %+v, want a flagged partial answer", groups)
	}
	out := renderText(t, SearchData{Query: "deploys", Groups: groups})
	if !strings.Contains(out, "most recent blocks only") {
		t.Fatalf("page does not say the walk stopped:\n%s", out)
	}
}

// A content search that filled its page says older deploys may match too.
func TestFullContentPageIsFlagged(t *testing.T) {
	t.Parallel()

	var txs []indexer.Tx
	for i := range recentLimit {
		var m indexer.Message
		m.Value.Type = "MsgAddPackage"
		m.Value.Package = &struct {
			Path string `json:"path"`
		}{Path: fmt.Sprintf("gno.land/r/alice/p%d", i)}
		txs = append(txs, indexer.Tx{Height: 100 - i, Messages: []indexer.Message{m}})
	}
	h := newHandler(t, &mockClient{}, &mockIndexer{txs: txs})
	groups, _ := h.Search(context.Background(), mustQuery(t, h, "content:avl.Tree", ""))
	if len(groups) != 1 || !groups[0].Truncated || !strings.Contains(groups[0].Notice, "newest") {
		t.Fatalf("groups = %+v, want the full page flagged", groups)
	}

	// A real failure is still a failure.
	h = newHandler(t, &mockClient{}, &mockIndexer{err: errors.New("boom")})
	groups, _ = h.Search(context.Background(), mustQuery(t, h, "content:avl.Tree", ""))
	if len(groups) != 1 || groups[0].Err == nil {
		t.Fatalf("groups = %+v, want a failed group", groups)
	}
}
