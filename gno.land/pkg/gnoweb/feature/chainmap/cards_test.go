package chainmap

import (
	"context"
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
		Doc:     "Package blog is a multi-author blog.\n\nAnyone can post.\n\n- one\n- two",
		Actions: []string{"Post"},
		Height:  25988,
	},
	"/r/demo/coins":  {Doc: "Package coins provides coin helpers.", Height: 0},
	"/r/demo/wugnot": {Actions: []string{"Approve", "Deposit", "Transfer", "TransferFrom", "Withdraw"}},
	"/p/demo/avl":    {Doc: "Package avl implements an AVL tree."},
}

func TestCards_FromChain(t *testing.T) {
	t.Parallel()

	h := New(Deps{Imports: fakeImports{}, Packages: fakePackages{summaries: cardSummaries}})
	cards := h.Cards(context.Background(), []string{"/r/demo/blog", "/r/demo/coins", "/r/demo/wugnot", "/p/demo/avl", "/r/demo/gone"})
	require.Len(t, cards, 5)

	assert.Equal(t, components.PackageCard{
		Path: "/r/demo/blog", Realm: true, Known: true,
		Description: "Multi-author blog. Anyone can post.",
		Actions:     1,
		Block:       25988,
	}, cards[0], "a realm with a call: no tag, its actions and its block")

	assert.Equal(t, "Provides coin helpers.", cards[1].Description)
	assert.Equal(t, []string{"read-only"}, cards[1].Tags, "a realm with nothing to call is read-only")
	assert.Zero(t, cards[1].Block, "deployed at genesis")
	assert.True(t, cards[1].Known)

	assert.Equal(t, []string{"token"}, cards[2].Tags, "a realm exposing the GRC20 calls is a token")
	assert.Equal(t, 5, cards[2].Actions)

	assert.Equal(t, components.PackageCard{
		Path: "/p/demo/avl", Known: true, Description: "Implements an AVL tree.",
	}, cards[3], "a pure package has no actions and no tags")

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

// A cold activity scan must not spend the listing's wait: the chain's part of
// each card still lands.
func TestCards_SlowIndexerDoesNotStarveCards(t *testing.T) {
	t.Parallel()

	idx := &fakeIndexer{tip: 1000, t0: time.Now().Add(-1000 * time.Second), bandDelay: 5 * time.Second}
	h := New(Deps{Indexer: idx, Imports: fakeImports{}, Packages: fakePackages{summaries: cardSummaries}, Domain: "gno.land"})
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	cards := h.Cards(ctx, []string{"/r/demo/blog"})
	assert.True(t, cards[0].Known, "the summary is read beside the scan, not after it")
	assert.Zero(t, cards[0].WeeklyUsers)
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

	cases := map[string]string{
		"":                                   "",
		"Package foo is a counter.":          "Counter.",
		"package foo implements counters.":   "Implements counters.",
		"Package foo is the main realm.":     "Main realm.",
		"Package bar provides helpers.":      "Provides helpers.",
		"Boards2 is a forum for discussion.": "Boards2 is a forum for discussion.",
	}
	for doc, want := range cases {
		assert.Equal(t, want, describe(doc), "doc %q", doc)
	}
}
