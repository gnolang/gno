package omnisearch

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/indexer"
	"github.com/gnolang/gno/gnovm/pkg/doc"
)

// --- mocks -----------------------------------------------------------------

type mockClient struct {
	doc     *doc.JSONDocumentation
	files   []string
	renders map[string]string
	docErr  error

	mu         sync.Mutex
	docCalls   int
	realmCalls int
}

func (m *mockClient) Realm(_ context.Context, path, _ string) ([]byte, error) {
	m.mu.Lock()
	m.realmCalls++
	m.mu.Unlock()

	body, ok := m.renders[path]
	if !ok {
		return nil, errors.New("no render")
	}
	return []byte(body), nil
}

func (m *mockClient) Doc(_ context.Context, _ string, _ int64) (*doc.JSONDocumentation, error) {
	m.docCalls++
	if m.docErr != nil {
		return nil, m.docErr
	}
	return m.doc, nil
}

func (m *mockClient) ListFiles(_ context.Context, _ string, _ int64) ([]string, error) {
	return m.files, nil
}

// mockDirectory stands in for the singleflight-backed RealmDirectory the
// wire-in supplies.
type mockDirectory struct {
	realms    []string
	packages  []string
	truncated bool
	err       error

	mu    sync.Mutex
	calls int
}

func (m *mockDirectory) Paths(context.Context) ([]string, []string, bool, error) {
	m.mu.Lock()
	m.calls++
	m.mu.Unlock()
	if m.err != nil {
		return nil, nil, false, m.err
	}
	return m.realms, m.packages, m.truncated, nil
}

type mockIndexer struct {
	tx  *indexer.Tx
	txs []indexer.Tx
	err error
}

func (m *mockIndexer) LatestBlockHeight(context.Context) (int, error) { return 185214, nil }
func (m *mockIndexer) URL() string                                    { return "https://indexer.example/graphql/query" }

func (m *mockIndexer) TxByHash(context.Context, string) (*indexer.Tx, error) {
	return m.tx, m.err
}

func (m *mockIndexer) RecentByPackage(context.Context, string, int) ([]indexer.Tx, error) {
	return m.txs, m.err
}

func (m *mockIndexer) RecentByAddress(context.Context, string, int) ([]indexer.Tx, error) {
	return m.txs, m.err
}

func (m *mockIndexer) Deploys(context.Context, string, int) ([]indexer.Tx, error) {
	return m.txs, m.err
}

func (m *mockIndexer) Importers(context.Context, string, int) ([]indexer.Tx, error) {
	return m.txs, m.err
}

func (m *mockIndexer) SourceContains(context.Context, string, int) ([]indexer.Tx, error) {
	return m.txs, m.err
}

func (m *mockIndexer) Block(context.Context, int) (*indexer.Block, error) {
	return &indexer.Block{Height: 185214, Hash: "abc", ChainID: "test"}, m.err
}

func newHandler(t *testing.T, c *mockClient, idx Indexer) *Handler {
	t.Helper()
	return newHandlerWithDir(t, c, &mockDirectory{}, idx)
}

func newHandlerWithDir(t *testing.T, c *mockClient, dir *mockDirectory, idx Indexer) *Handler {
	t.Helper()
	deps := Deps{
		Client:    c,
		Directory: dir,
		Domain:    "gno.land",
		Logger:    slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	// Assigned inside the branch so a nil *mockIndexer never reaches the
	// interface field as a non-nil interface — the same trap the wire-in
	// guards against.
	if idx != nil {
		deps.Indexer = idx
	}
	return New(deps)
}

func mustQuery(t *testing.T, h *Handler, raw, pkgPath string) *Query {
	t.Helper()
	q, err := ParseQuery(raw)
	if err != nil {
		t.Fatalf("ParseQuery(%q): %v", raw, err)
	}
	if pkgPath != "" {
		q.PkgPath = pkgPath
		q.ChainPath = "gno.land" + pkgPath
	}
	return q
}

// --- the feature switch ----------------------------------------------------

// With no indexer configured, the indexer-backed qualifiers must not exist:
// not in the hint list, and not answerable. A reader who types one is told it
// is not a qualifier here, rather than being shown an empty result list that
// implies the transaction does not exist.
func TestWithoutIndexerTheIndexerQualifiersDoNotExist(t *testing.T) {
	t.Parallel()

	h := newHandler(t, &mockClient{}, nil)

	for _, sel := range h.selectors {
		if sel.Source == SourceIndexer {
			t.Errorf("selector %q is registered without an indexer", sel.Name)
		}
	}

	groups, unknown := h.Search(context.Background(), mustQuery(t, h, "tx:9f2a", "/r/demo/boards"))
	if unknown != "tx" {
		t.Fatalf("unknown = %q, want %q", unknown, "tx")
	}
	if groups != nil {
		t.Fatalf("groups = %+v, want none", groups)
	}
}

func TestWithIndexerTheIndexerQualifiersAnswer(t *testing.T) {
	t.Parallel()

	idx := &mockIndexer{tx: &indexer.Tx{Hash: "9f2a", Height: 185214, Success: true, GasUsed: 12340}}
	h := newHandler(t, &mockClient{}, idx)

	groups, unknown := h.Search(context.Background(), mustQuery(t, h, "tx:9f2a", ""))
	if unknown != "" {
		t.Fatalf("unknown = %q, want none", unknown)
	}
	if len(groups) != 1 || len(groups[0].Results) != 1 {
		t.Fatalf("groups = %+v, want one group with one result", groups)
	}
	if groups[0].Source != SourceIndexer {
		t.Errorf("source = %q, want %q", groups[0].Source, SourceIndexer)
	}
}

// --- failure handling ------------------------------------------------------

// A resolver that fails must degrade to a group that says so. An indexer
// being down is not a reason to fail the page.
func TestResolverFailureBecomesAVisibleGroupError(t *testing.T) {
	t.Parallel()

	h := newHandler(t, &mockClient{}, &mockIndexer{err: errors.New("indexer unreachable")})

	groups, _ := h.Search(context.Background(), mustQuery(t, h, "activity", "/r/demo/boards"))
	if len(groups) != 1 {
		t.Fatalf("groups = %+v, want one", groups)
	}
	if groups[0].Err == nil {
		t.Fatal("group Err = nil, want the resolver failure")
	}
}

// A package-scoped qualifier typed on a path that names no package must say
// what is missing rather than querying a path the chain does not hold.
func TestPackageScopedQualifierNeedsAPackage(t *testing.T) {
	t.Parallel()

	c := &mockClient{}
	h := newHandler(t, c, nil)

	groups, _ := h.Search(context.Background(), mustQuery(t, h, "func:Vote", ""))
	if len(groups) != 1 || groups[0].Err == nil {
		t.Fatalf("groups = %+v, want one group carrying an error", groups)
	}
	if c.docCalls != 0 {
		t.Errorf("Doc called %d times, want 0 — nothing should reach the chain", c.docCalls)
	}
}

func TestShortTermIsRefusedBeforeAnyFetch(t *testing.T) {
	t.Parallel()

	c := &mockClient{}
	h := newHandler(t, c, nil)

	groups, _ := h.Search(context.Background(), mustQuery(t, h, "func:V", "/r/demo/boards"))
	if len(groups) != 1 || groups[0].Err == nil {
		t.Fatalf("groups = %+v, want one group carrying an error", groups)
	}
	if c.docCalls != 0 {
		t.Errorf("Doc called %d times, want 0", c.docCalls)
	}
}

// --- chain-backed resolvers ------------------------------------------------

func TestFuncQualifierAnswersFromQdoc(t *testing.T) {
	t.Parallel()

	c := &mockClient{doc: &doc.JSONDocumentation{
		Funcs: []*doc.JSONFunc{
			{Name: "Vote", Signature: "func Vote(cur realm, id int)", Crossing: true, File: "vote.gno", Line: 42},
			{Name: "Render", Signature: "func Render(path string) string", File: "render.gno", Line: 7},
		},
	}}
	h := newHandler(t, c, nil)

	groups, _ := h.Search(context.Background(), mustQuery(t, h, "func:Vote", "/r/demo/boards"))
	if len(groups) != 1 || len(groups[0].Results) != 1 {
		t.Fatalf("groups = %+v, want one group with one result", groups)
	}
	got := groups[0].Results[0]
	if got.Href != "/r/demo/boards$file=vote.gno&source#L42" {
		t.Errorf("Href = %q", got.Href)
	}
	if !slices.Contains(got.Tags, "crossing") {
		t.Errorf("Tags = %v, want to include crossing", got.Tags)
	}
}

// --- discovery -------------------------------------------------------------

func newDiscoveryDir() *mockDirectory {
	return &mockDirectory{
		realms: []string{
			"gno.land/r/alice/blog",
			"gno.land/r/alice/shop",
			"gno.land/r/bob/blog",
		},
		packages: []string{"gno.land/p/alice/util"},
	}
}

func newDiscoveryClient() *mockClient { return &mockClient{} }

func TestDiscoveryGroupsRealmsPackagesAndUsers(t *testing.T) {
	t.Parallel()

	dir := newDiscoveryDir()
	h := newHandlerWithDir(t, newDiscoveryClient(), dir, nil)

	groups, _ := h.Search(context.Background(), mustQuery(t, h, "blog", ""))

	labels := map[string]int{}
	for _, g := range groups {
		labels[g.Label] = len(g.Results)
	}
	if labels["Realms"] != 2 {
		t.Errorf("Realms = %d, want 2", labels["Realms"])
	}
	if labels["Packages"] != 0 {
		t.Errorf("Packages = %d, want 0", labels["Packages"])
	}
	// The users group is derived from the matched paths, so it costs no fetch.
	if labels["Users"] != 2 {
		t.Errorf("Users = %d, want 2 (alice, bob)", labels["Users"])
	}
	// One directory call, coalesced with every other caller asking the same
	// thing: the discovery search is the widest thing the omnibar does.
	if dir.calls != 1 {
		t.Errorf("Directory.Paths called %d times, want 1", dir.calls)
	}
}

func TestDiscoveryNarrowsByAuthor(t *testing.T) {
	t.Parallel()

	h := newHandlerWithDir(t, newDiscoveryClient(), newDiscoveryDir(), nil)

	groups, _ := h.Search(context.Background(), mustQuery(t, h, "author:alice", ""))
	for _, g := range groups {
		for _, r := range g.Results {
			if g.Label == "Users" {
				continue
			}
			if !strings.Contains(r.Title, "/alice/") {
				t.Errorf("%s: %q leaked past author:alice", g.Label, r.Title)
			}
		}
	}
}

func TestDiscoveryNarrowsByKind(t *testing.T) {
	t.Parallel()

	h := newHandlerWithDir(t, newDiscoveryClient(), newDiscoveryDir(), nil)

	groups, _ := h.Search(context.Background(), mustQuery(t, h, "blog is:realm", ""))
	for _, g := range groups {
		if g.Label == "Packages" {
			t.Error("is:realm still listed a Packages group")
		}
	}
}

// --- provenance ------------------------------------------------------------

// The freshness footer exists for indexer results. A chain-only answer must
// not reach out to the indexer just to stamp a footer nothing will show.
func TestIndexerStatusOnlyWhenAnIndexerGroupRan(t *testing.T) {
	t.Parallel()

	h := newHandlerWithDir(t, newDiscoveryClient(), newDiscoveryDir(), &mockIndexer{})

	chainOnly := h.build(context.Background(), mustQuery(t, h, "blog", ""))
	if chainOnly.Indexer != nil {
		t.Errorf("Indexer footer = %+v on a chain-only answer, want nil", chainOnly.Indexer)
	}

	withIndexer := h.build(context.Background(), mustQuery(t, h, "account:g1abc", ""))
	if withIndexer.Indexer == nil {
		t.Fatal("Indexer footer = nil on an indexer-backed answer")
	}
	if withIndexer.Indexer.LastBlock != 185214 {
		t.Errorf("LastBlock = %d, want 185214", withIndexer.Indexer.LastBlock)
	}
}

// --- rendered-content search -----------------------------------------------

// Rendering costs a node execution per realm, so the search refuses to run
// until the field is narrowed. Without this it would be the amplification
// ADR-003 §Resource bounds exists to prevent.
func TestRenderSearchRefusesAnUnboundedField(t *testing.T) {
	t.Parallel()

	c := newDiscoveryClient()
	h := newHandlerWithDir(t, c, newDiscoveryDir(), nil)

	groups, _ := h.Search(context.Background(), mustQuery(t, h, "render:hello", ""))
	if len(groups) != 1 || groups[0].Err == nil {
		t.Fatalf("groups = %+v, want one group carrying an error", groups)
	}
	if c.realmCalls != 0 {
		t.Errorf("Realm called %d times, want 0", c.realmCalls)
	}
}

func TestRenderSearchFindsTextInRenderedOutput(t *testing.T) {
	t.Parallel()

	c := newDiscoveryClient()
	c.renders = map[string]string{
		"/r/alice/blog": "# Alice's blog\n\nA post about proposals and voting.",
		"/r/alice/shop": "# Shop\n\nNothing relevant here.",
	}
	h := newHandlerWithDir(t, c, newDiscoveryDir(), nil)

	groups, _ := h.Search(context.Background(), mustQuery(t, h, "render:proposals author:alice", ""))
	if len(groups) != 1 {
		t.Fatalf("groups = %+v, want one", groups)
	}
	if groups[0].Err != nil {
		t.Fatalf("group Err = %v", groups[0].Err)
	}
	if len(groups[0].Results) != 1 {
		t.Fatalf("results = %+v, want the one realm whose render matches", groups[0].Results)
	}
	got := groups[0].Results[0]
	if got.Title != "/r/alice/blog" {
		t.Errorf("Title = %q, want /r/alice/blog", got.Title)
	}
	if !strings.Contains(got.Detail, "proposals") {
		t.Errorf("Detail = %q, want a snippet around the match", got.Detail)
	}
	// Rendered output comes from the node executing Render, so it is chain
	// data and must not be marked as an indexer's best effort.
	if groups[0].Source != SourceChain {
		t.Errorf("Source = %q, want %q", groups[0].Source, SourceChain)
	}
}

// `in:` narrows to a single realm, which is the cheapest possible field.
func TestRenderSearchScopedByInRendersOnlyThatRealm(t *testing.T) {
	t.Parallel()

	c := newDiscoveryClient()
	c.renders = map[string]string{"/r/alice/blog": "voting is open"}
	dir := newDiscoveryDir()
	h := newHandlerWithDir(t, c, dir, nil)

	q := mustQuery(t, h, "render:voting in:/r/alice/blog", "")
	h.scope(q, parseURL(t, "/$search"))

	groups, _ := h.Search(context.Background(), q)
	if len(groups) != 1 || len(groups[0].Results) != 1 {
		t.Fatalf("groups = %+v, want one result", groups)
	}
	if c.realmCalls != 1 {
		t.Errorf("Realm called %d times, want 1", c.realmCalls)
	}
	if dir.calls != 0 {
		t.Errorf("Directory.Paths called %d times, want 0 — in: names the realm outright", dir.calls)
	}
}

// A realm that fails to render is skipped, not fatal: the others still have
// answers, and a realm without a Render is not a search error.
func TestRenderSearchSkipsRealmsThatDoNotRender(t *testing.T) {
	t.Parallel()

	c := newDiscoveryClient()
	c.renders = map[string]string{"/r/alice/shop": "voting"}
	h := newHandlerWithDir(t, c, newDiscoveryDir(), nil)

	groups, _ := h.Search(context.Background(), mustQuery(t, h, "render:voting author:alice", ""))
	if len(groups) != 1 || groups[0].Err != nil {
		t.Fatalf("groups = %+v, want one clean group", groups)
	}
	if len(groups[0].Results) != 1 {
		t.Fatalf("results = %d, want the one realm that rendered", len(groups[0].Results))
	}
}

func TestRenderSearchCapsTheField(t *testing.T) {
	t.Parallel()

	paths := make([]string, 0, maxRenderCandidates*3)
	renders := make(map[string]string, maxRenderCandidates*3)
	for i := range maxRenderCandidates * 3 {
		p := "gno.land/r/alice/app" + string(rune('a'+i))
		paths = append(paths, p)
		renders[strings.TrimPrefix(p, "gno.land")] = "voting"
	}
	c := &mockClient{renders: renders}
	h := newHandlerWithDir(t, c, &mockDirectory{realms: paths}, nil)

	if _, _ = h.Search(context.Background(), mustQuery(t, h, "render:voting author:alice", "")); c.realmCalls > maxRenderCandidates {
		t.Fatalf("Realm called %d times, want at most %d", c.realmCalls, maxRenderCandidates)
	}
}

// Typing "alice/" lists that namespace's other packages — the sibling
// suggestion the omnibar mock-up shows — without a dedicated code path: the
// discovery substring match already covers it.
func TestNamespaceSlashListsSiblings(t *testing.T) {
	t.Parallel()

	h := newHandlerWithDir(t, newDiscoveryClient(), newDiscoveryDir(), nil)

	groups, _ := h.Search(context.Background(), mustQuery(t, h, "alice/", ""))

	var titles []string
	for _, g := range groups {
		if g.Label == "Users" {
			continue
		}
		for _, r := range g.Results {
			titles = append(titles, r.Title)
		}
	}
	want := map[string]bool{"/r/alice/blog": true, "/r/alice/shop": true, "/p/alice/util": true}
	if len(titles) != len(want) {
		t.Fatalf("titles = %v, want alice's three packages", titles)
	}
	for _, got := range titles {
		if !want[got] {
			t.Errorf("%q is not one of alice's packages", got)
		}
	}
}

// The costly selectors carry their own floor. A two-character `content:` is
// not a search, it is a substring scan of every deployed file body on the
// chain — and gnoweb is the amplifier even though the CPU is the indexer's.
func TestExpensiveSelectorsRaiseTheirOwnMinimum(t *testing.T) {
	t.Parallel()

	tests := []struct {
		selector string
		want     int
	}{
		{"content", 4},
		{"render", 3},
		{"func", MinTermLen},
	}

	h := newHandlerWithDir(t, newDiscoveryClient(), newDiscoveryDir(), &mockIndexer{})
	for _, tt := range tests {
		t.Run(tt.selector, func(t *testing.T) {
			t.Parallel()

			sel, ok := h.byName[tt.selector]
			if !ok {
				t.Fatalf("selector %q is not registered", tt.selector)
			}
			if got := sel.minTerm(); got != tt.want {
				t.Fatalf("%s minTerm = %d, want %d", tt.selector, got, tt.want)
			}
		})
	}
}

// Fan-out is a property of the selector, so the gate has to be on the
// selector — not a special case buried in the one resolver that is expensive
// today.
func TestOnlyFanoutSelectorsArePageOnly(t *testing.T) {
	t.Parallel()

	h := newHandlerWithDir(t, newDiscoveryClient(), newDiscoveryDir(), &mockIndexer{})

	pageOnly := map[string]bool{}
	for _, sel := range h.selectors {
		if sel.PageOnly {
			pageOnly[sel.Name] = true
		}
	}
	// Every selector that fans out to a third party, or scans the whole
	// chain, belongs here. An earlier version of this test asserted
	// {render} alone and so enshrined content:/importers running their
	// full-chain substring scan once per keystroke.
	want := map[string]bool{"render": true, "content": true, "importers": true}
	if len(pageOnly) != len(want) {
		t.Fatalf("page-only selectors = %v, want %v", pageOnly, want)
	}
	for name := range want {
		if !pageOnly[name] {
			t.Errorf("%s fans out but is not page-only", name)
		}
	}
}

// A bare selector takes no argument, but a reader who types `activity:foo`
// plainly wants activity. Before, it matched neither the selector lookup nor
// the unknown-qualifier check and fell through to a path search with empty
// text — a blank page explaining nothing.
func TestBareSelectorGivenAValueStillAnswers(t *testing.T) {
	t.Parallel()

	idx := &mockIndexer{txs: []indexer.Tx{{Hash: "abc", Height: 10, Success: true}}}
	h := newHandlerWithDir(t, newDiscoveryClient(), newDiscoveryDir(), idx)

	groups, unknown := h.Search(context.Background(),
		mustQuery(t, h, "activity:foo", "/r/demo/boards"))

	if unknown != "" {
		t.Fatalf("unknown = %q, want none", unknown)
	}
	if len(groups) != 1 || groups[0].Label != "Recent activity" {
		t.Fatalf("groups = %+v, want the activity group", groups)
	}
	if len(groups[0].Results) != 1 {
		t.Fatalf("results = %d, want the selector to have run", len(groups[0].Results))
	}
}

// Resolvers pre-size against the whole package (len(jdoc.Funcs) and
// friends), so a reslice would keep the whole array alive for as long as a
// twenty-row page renders.
//
// Non-aliasing is the property under test, not capacity: `rs[:20:20]` reports
// a capacity of 20 while still pointing into the original allocation, and the
// GC frees allocations, not slices. A copy cannot retain it; an alias can.
func TestCapResultsCopiesRatherThanReslicing(t *testing.T) {
	t.Parallel()

	oversized := make([]Result, 0, 5000)
	for i := range MaxResults * 10 {
		oversized = append(oversized, Result{Title: strconv.Itoa(i)})
	}

	got := capResults(oversized)
	if len(got) != MaxResults {
		t.Fatalf("len = %d, want %d", len(got), MaxResults)
	}

	oversized[0].Title = "mutated"
	if got[0].Title == "mutated" {
		t.Fatal("the capped slice aliases its source, so it pins the full backing array")
	}
}

// The node caps qpaths and always drops the same lexicographic tail, so a
// capped listing must say so rather than report "no match".
func TestTruncatedListingIsReportedNotSwallowed(t *testing.T) {
	t.Parallel()

	dir := newDiscoveryDir()
	dir.truncated = true
	h := newHandlerWithDir(t, newDiscoveryClient(), dir, nil)

	groups, _ := h.Search(context.Background(), mustQuery(t, h, "blog", ""))
	if len(groups) == 0 {
		t.Fatal("no groups")
	}
	for _, g := range groups {
		if !g.Truncated {
			t.Errorf("group %q does not carry the truncation flag", g.Label)
		}
	}
}
