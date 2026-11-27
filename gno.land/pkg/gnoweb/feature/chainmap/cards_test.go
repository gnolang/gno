package chainmap

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/indexer"
)

// fakePackages answers summaries from a map; a path absent from it is not
// live. delay makes every read take that long, or until its ctx ends.
type fakePackages struct {
	summaries map[string]PackageSummary
	delay     time.Duration
}

func (f fakePackages) Summary(ctx context.Context, pkgPath string) (PackageSummary, error) {
	if f.delay > 0 {
		select {
		case <-time.After(f.delay):
		case <-ctx.Done():
			return PackageSummary{}, ctx.Err()
		}
	}
	s, ok := f.summaries[pkgPath]
	if !ok {
		return PackageSummary{}, ErrNotLive
	}
	return s, nil
}

var cardSummaries = map[string]PackageSummary{
	"/r/demo/blog": {
		Name:    "blog",
		Doc:     "Package blog is a multi-author blog.\n\nAnyone can post.\n\n- one\n- two",
		Actions: []string{"Post"},
		Height:  25988,
	},
	"/r/demo/coins":  {Name: "coins", Doc: "Package coins provides coin helpers.", Height: 0},
	"/r/demo/wugnot": {Name: "wugnot", Actions: []string{"Approve", "Deposit", "Transfer", "TransferFrom", "Withdraw"}},
	"/p/demo/avl":    {Name: "avl", Doc: "Package avl implements an AVL tree."},
}

func TestCards_FromChain(t *testing.T) {
	t.Parallel()

	h := New(Deps{Imports: fakeImports{}, Packages: fakePackages{summaries: cardSummaries}})
	cards := h.Cards(context.Background(), []string{"/r/demo/blog", "/r/demo/coins", "/r/demo/wugnot", "/p/demo/avl", "/r/demo/gone"})
	require.Len(t, cards, 5)

	assert.Equal(t, components.PackageCard{
		Path: "/r/demo/blog", Realm: true, Known: true,
		Description: "multi-author blog. Anyone can post.",
		Actions:     1,
		Block:       25988,
	}, cards[0], "a realm with a call: its actions and its block")

	assert.Equal(t, components.PackageCard{
		Path: "/r/demo/coins", Realm: true, Known: true, Description: "provides coin helpers.",
	}, cards[1], "a realm with nothing to call, deployed at genesis")

	assert.Equal(t, 5, cards[2].Actions)

	assert.Equal(t, components.PackageCard{
		Path: "/p/demo/avl", Known: true, Description: "implements an AVL tree.",
	}, cards[3], "a pure package has no actions")

	assert.Equal(t, components.PackageCard{Path: "/r/demo/gone", Realm: true}, cards[4],
		"a package the chain does not know keeps its path and links only")
}

func TestCards_WithoutReader(t *testing.T) {
	t.Parallel()

	cards := New(Deps{Imports: fakeImports{}}).Cards(context.Background(), []string{"/r/demo/blog"})
	assert.Equal(t, []components.PackageCard{{Path: "/r/demo/blog", Realm: true}}, cards)
}

func TestCards_SlowChainLeavesCardsBlank(t *testing.T) {
	t.Parallel()

	h := New(Deps{Imports: fakeImports{}, Packages: fakePackages{summaries: cardSummaries, delay: time.Second}})
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	start := time.Now()
	cards := h.Cards(ctx, []string{"/r/demo/blog"})
	assert.Less(t, time.Since(start), 500*time.Millisecond, "the listing does not wait for the chain")
	assert.Equal(t, []components.PackageCard{{Path: "/r/demo/blog", Realm: true}}, cards)
}

func TestCards_WithIndexer(t *testing.T) {
	t.Parallel()

	const tip = 1000
	t0 := time.Now().Add(-tip * time.Second)
	idx := &fakeIndexer{tip: tip, t0: t0, calls: map[int][]indexer.Tx{
		900: {call(900, true, "g1alice", "gno.land/r/demo/blog")},
		950: {call(950, true, "g1bob", "gno.land/r/demo/blog"), call(950, false, "g1eve", "gno.land/r/demo/blog")},
	}}
	pkgs := map[string]PackageSummary{
		"/r/demo/blog":  {Height: 500},
		"/r/demo/quiet": {Height: 0},
	}
	h := New(Deps{Indexer: idx, Imports: fakeImports{}, Packages: fakePackages{summaries: pkgs}, Domain: "gno.land"})
	_, err := h.weekActivity(context.Background())
	require.NoError(t, err)

	cards := h.Cards(context.Background(), []string{"/r/demo/blog", "/r/demo/quiet"})

	blog := cards[0]
	assert.Equal(t, 2, blog.WeeklyUsers, "distinct successful callers")
	assert.False(t, blog.AtLeast)
	assert.WithinDuration(t, t0.Add(500*time.Second), blog.Date, time.Second, "the deploy block's time")
	assert.WithinDuration(t, t0.Add(950*time.Second), blog.LastUsed, 5*time.Second, "about when block 950 ran")

	quiet := cards[1]
	assert.Zero(t, quiet.WeeklyUsers, "an uncalled realm shows no activity")
	assert.True(t, quiet.LastUsed.IsZero())
	assert.True(t, quiet.Date.IsZero(), "genesis has no block to date")
}

// A listing never waits on the activity scan: with an indexer that never
// answers, the chain's part of each card lands and the listing returns as
// soon as it has, while the scan it started carries on.
func TestCards_SlowIndexerDoesNotStarveCards(t *testing.T) {
	t.Parallel()

	idx := &fakeIndexer{tip: 1000, t0: time.Now().Add(-1000 * time.Second), bandDelay: 5 * time.Second}
	h := New(Deps{Indexer: idx, Imports: fakeImports{}, Packages: fakePackages{summaries: cardSummaries}, Domain: "gno.land"})

	for range 3 {
		start := time.Now()
		cards := h.Cards(context.Background(), []string{"/r/demo/blog"})
		assert.Less(t, time.Since(start), 300*time.Millisecond, "the listing does not wait on the scan")
		assert.True(t, cards[0].Known, "the summary is read beside the scan, not after it")
		assert.Zero(t, cards[0].WeeklyUsers)
	}
	assert.Eventually(t, func() bool {
		idx.mu.Lock()
		defer idx.mu.Unlock()
		return len(idx.bandsQueried) > 0
	}, time.Second, 10*time.Millisecond, "the scan was started for the next reader")
}

// An aggregate whose tip is long past means the indexer stopped answering:
// "over the last 7 days" would no longer hold, so the cards leave it off.
func TestCards_StaleActivityIsLeftOff(t *testing.T) {
	t.Parallel()

	idx := &fakeIndexer{tip: 1000, t0: time.Now().Add(-1000 * time.Second)}
	h := New(Deps{Indexer: idx, Imports: fakeImports{}, Packages: fakePackages{summaries: cardSummaries}, Domain: "gno.land"})
	old := time.Now().Add(-10 * 24 * time.Hour)
	h.activity.store(activityKey, &Activity{
		Callers: map[string]int{"gno.land/r/demo/blog": 2},
		Last:    map[string]int{"gno.land/r/demo/blog": 900},
		From:    1, To: 1000, Since: old.Add(-7 * 24 * time.Hour), Until: old,
	}, nil)

	cards := h.Cards(context.Background(), []string{"/r/demo/blog"})
	assert.True(t, cards[0].Known)
	assert.Zero(t, cards[0].WeeklyUsers)
	assert.True(t, cards[0].LastUsed.IsZero())
}

// panickingPackages panics on every read, as a doc comment that trips the
// markdown parser would.
type panickingPackages struct{}

func (panickingPackages) Summary(context.Context, string) (PackageSummary, error) {
	panic("index out of range")
}

// A read that panics runs on the flight's goroutine, out of reach of
// net/http's recover: it must leave a blank card, not stop gnoweb.
func TestCards_PanickingReadLeavesTheCardBlank(t *testing.T) {
	t.Parallel()

	h := New(Deps{Imports: fakeImports{}, Packages: panickingPackages{}})
	for range 2 {
		cards := h.Cards(context.Background(), []string{"/r/demo/blog"})
		assert.Equal(t, []components.PackageCard{{Path: "/r/demo/blog", Realm: true}}, cards)
	}
	assert.Eventually(t, func() bool { return len(h.cardSlots) == 0 }, time.Second, 10*time.Millisecond, "the slot is given back")
}

// Pure packages are never called: their listing does not ask for activity.
func TestCards_PureListingSkipsActivity(t *testing.T) {
	t.Parallel()

	idx := &fakeIndexer{tip: 1000, t0: time.Now().Add(-1000 * time.Second)}
	h := New(Deps{Indexer: idx, Imports: fakeImports{}, Packages: fakePackages{summaries: cardSummaries}, Domain: "gno.land"})

	h.Cards(context.Background(), []string{"/p/demo/avl"})
	idx.mu.Lock()
	defer idx.mu.Unlock()
	assert.Empty(t, idx.bandsQueried)
}

// With every fetch slot taken, a card stays blank and nothing is remembered:
// the next reader gets the answer.
func TestCards_BusySlotsAreNotRemembered(t *testing.T) {
	t.Parallel()

	h := New(Deps{Imports: fakeImports{}, Packages: fakePackages{summaries: cardSummaries}})
	for range maxCardFetches {
		h.cardSlots <- struct{}{}
	}
	assert.False(t, h.Cards(context.Background(), []string{"/r/demo/blog"})[0].Known)

	for range maxCardFetches {
		<-h.cardSlots
	}
	assert.True(t, h.Cards(context.Background(), []string{"/r/demo/blog"})[0].Known)
}

// countingPackages answers every path, slowly, and records the most reads
// it ever had in flight.
type countingPackages struct {
	mu            sync.Mutex
	inFlight, max int
	calls         int
}

func (c *countingPackages) Summary(ctx context.Context, pkgPath string) (PackageSummary, error) {
	c.mu.Lock()
	c.inFlight++
	c.calls++
	c.max = max(c.max, c.inFlight)
	c.mu.Unlock()
	defer func() {
		c.mu.Lock()
		c.inFlight--
		c.mu.Unlock()
	}()
	select {
	case <-time.After(50 * time.Millisecond):
	case <-ctx.Done():
	}
	return PackageSummary{Height: 7}, nil
}

// A flood of cold listings, each over its own packages, never runs more than
// maxCardFetches backend reads at once, chain and indexer together, and never
// waits past its budget.
func TestCards_FloodStaysUnderTheCap(t *testing.T) {
	t.Parallel()

	pkgs := &countingPackages{}
	idx := &fakeIndexer{tip: 1000, t0: time.Now().Add(-1000 * time.Second)}
	h := New(Deps{Indexer: idx, Imports: fakeImports{}, Packages: pkgs, Domain: "gno.land"})

	const listings, perListing = 40, 50
	var wg sync.WaitGroup
	start := time.Now()
	for l := range listings {
		paths := make([]string, perListing)
		for i := range paths {
			paths[i] = fmt.Sprintf("/p/flood%d/pkg%d", l, i)
		}
		wg.Go(func() {
			ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
			defer cancel()
			h.Cards(ctx, paths)
		})
	}
	wg.Wait()
	assert.Less(t, time.Since(start), 2*time.Second, "every listing returns within its budget")

	// Fetches outlive their listing; once they land, every slot is back.
	assert.Eventually(t, func() bool { return len(h.cardSlots) == 0 }, 2*time.Second, 10*time.Millisecond)
	pkgs.mu.Lock()
	defer pkgs.mu.Unlock()
	assert.LessOrEqual(t, pkgs.max, maxCardFetches, "chain reads in flight")
	assert.Less(t, pkgs.calls, listings*perListing, "refused reads are not run")
}

// A cached aggregate dates calls from the chain's own clock, not the reader's.
func TestActivity_TimeOf(t *testing.T) {
	t.Parallel()

	since := time.Unix(1_000, 0)
	a := &Activity{From: 100, To: 200, Since: since, Until: since.Add(100 * time.Second)}
	assert.Equal(t, since.Add(50*time.Second), a.timeOf(150))
	assert.Equal(t, a.Until, a.timeOf(200))
}

func TestDescribe(t *testing.T) {
	t.Parallel()

	cases := []struct{ name, doc, want string }{
		{"foo", "", ""},
		{"foo", "Package foo is a counter.", "counter."},
		{"foo", "package foo implements counters.", "implements counters."},
		{"foo", "Package foo is the main realm.", "main realm."},
		{"home", "Package home is the gno.land home page.", "gno.land home page."},
		// Only the package's own name makes the clause.
		{"hub", "Package registry for the chain.", "Package registry for the chain."},
		{"x", "Package management made easy for realms.", "Package management made easy for realms."},
		// The author's case is kept.
		{"foo20", "foo20 is a GRC20 token contract.", "foo20 is a GRC20 token contract."},
		{"boards2", "Boards2 is a forum for discussion.", "Boards2 is a forum for discussion."},
		// A disclaimer first: the card starts at the paragraph naming the
		// package.
		{"avl", "v0 - Unaudited: use at your own risk.\n\nPackage avl provides an AVL tree.\n\nMore.", "provides an AVL tree. More."},
		// No clause anywhere: the doc comment from its start.
		{"avl", "v0 - Unaudited.\n\nAn AVL tree.", "v0 - Unaudited. An AVL tree."},
		// A package qdoc could not name.
		{"", "Package foo is a counter.", "Package foo is a counter."},
	}
	for _, c := range cases {
		assert.Equal(t, c.want, describe(c.doc, c.name), "doc %q of %q", c.doc, c.name)
	}
}
