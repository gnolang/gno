package store

import (
	"sort"

	"gno.land/p/nt/avl/pager/v0"
	"gno.land/p/nt/avl/v0"
)

const (
	builderRank    = 10       // builders per ranking
	rediscoverSize = gridPool // Rediscover is not pinned: room for what is shown above
	rediscoverStep = rediscoverSize / 2
	pageSize       = 24
	minTrending    = 2   // stars_7d needed to trend
	trendingScan   = 200 // apps read at most for trending
)

// shelf is one section of the store front or the build lens, drawn as a
// grid. The realm decides what it holds and in which order; gnoweb draws it.
type shelf struct {
	Title  string
	Note   string // one line under the title, if any
	Items  []*listing
	Counts []int // per item, the ranked stars of the last 7 epochs (Trending)
	More   list  // the list behind the shelf, if any (Key != "")
	// Pinned shelves are the front's fixed sections: always drawn, from
	// their true head (only the Spotlight's apps are set aside), with Empty
	// when they hold no app.
	Pinned bool
	Empty  string
}

// list is a list of the store front, the "see all" of a shelf: paged over
// an index, or computed, one page long (Trending).
type list struct {
	Key    string
	Title  string
	Index  *avl.Tree  // paged lists
	Items  []*listing // computed lists
	Counts []int      // computed lists: each item's ranked stars this week, which its shelf shows
}

// findList returns the list with this key at height h.
func findList(key string, h int64) (list, bool) {
	switch key {
	case "top":
		return list{Key: key, Title: "Top", Index: apps.stars}, true
	case "latest":
		return list{Key: key, Title: "New", Index: apps.created}, true
	case "updated":
		return list{Key: key, Title: "Recently updated", Index: apps.updated}, true
	case "trending":
		t := trending(h, pageSize)
		return list{Key: key, Title: "Trending this week", Items: t.items, Counts: t.scores}, true
	}
	return list{}, false
}

func mustList(key string, h int64) list {
	l, ok := findList(key, h)
	if !ok {
		panic("store: unknown list " + key)
	}
	return l
}

// page returns page n (from 1) of the list, and its number of pages.
func (l list) page(n int) ([]*listing, int) {
	if l.Index != nil {
		return pageOf(l.Index, n)
	}
	if n != 1 {
		return nil, 1
	}
	return l.Items, 1
}

func (l list) count() int {
	if l.Index != nil {
		return l.Index.Size()
	}
	return len(l.Items)
}

// buildShelves computes the build lens shelves, over services and packages,
// for api/v1/build and the markdown build page.
func buildShelves(h int64) []shelf {
	return []shelf{
		{Title: "New to build with", Items: takeWhile(toolbox.created, gridPool, func(l *listing) bool { return l.isNew(h) })},
		{Title: "Recently updated", Items: window(toolbox.updated, 0, gridPool)},
		{Title: "Most starred", Items: window(toolbox.stars, 0, gridPool)},
	}
}

// takeWhile returns the first k listings of an index, stopping early at the
// first one for which while is false.
func takeWhile(index *avl.Tree, k int, while func(*listing) bool) []*listing {
	var out []*listing
	index.Iterate("", "", func(_ string, v any) bool {
		l := v.(*listing)
		if !while(l) {
			return true
		}
		out = append(out, l)
		return len(out) >= k
	})
	return out
}

// window returns the listings of an index from offset, at most n.
func window(index *avl.Tree, offset, n int) []*listing {
	var out []*listing
	index.IterateByOffset(offset, n, func(_ string, v any) bool {
		out = append(out, v.(*listing))
		return false
	})
	return out
}

// pageOf returns page n (from 1) of an index, and the number of pages; an
// empty index has one empty page.
func pageOf(index *avl.Tree, n int) ([]*listing, int) {
	p := pager.NewPager(index, pageSize, false).GetPage(n)
	out := make([]*listing, len(p.Items))
	for i, item := range p.Items {
		out[i] = item.Value.(*listing)
	}
	if p.TotalPages == 0 {
		return out, 1
	}
	return out, p.TotalPages
}

// trending ranks at most n apps by ranked stars over the trending window,
// with their scores. Each live
// slot of the trend index orders its apps by a score that can only have
// decayed since it was filed: reading a slot downwards, once a stored score
// is below minTrending or the n-th best real score, no later app of that
// slot can enter, so the result is exact unless trendingScan, which bounds
// the reads whatever the store's size, cuts it short.
func trending(h int64, n int) *topN {
	best, scanned, epoch := &topN{n: n}, 0, h/epochBlocks
	for _, s := range appTrend.slots {
		if s.byScore == nil || epoch-s.epoch >= trendEpochs {
			continue
		}
		s.byScore.Iterate("", "", func(key string, v any) bool {
			bound := int(descValue(key))
			if bound < minTrending || len(best.items) == best.n && bound < best.scores[best.n-1] {
				return true
			}
			l := v.(*listing)
			if n := l.stars7d(h); n >= minTrending {
				best.offer(n, l)
			}
			scanned++
			return scanned >= trendingScan
		})
		if scanned >= trendingScan {
			break
		}
	}
	return best
}

// rotation shows a window of rediscoverSize apps older than newWindow,
// which slides by rediscoverStep every epoch. gnoweb draws only the first
// half of a window once it sets aside what is shown above, so the step is
// half the window: every app falls in the first half of a window within
// ceil(m/rediscoverStep) epochs, with no state, no randomness and no pass
// over the apps. It rotates over named namespaces only: a free address
// namespace earns no guaranteed front-page slot, its apps stay in their
// category and the lists. The core apps stay in gno.land essentials. The
// index is newest first, so the new apps are a prefix, found by binary
// search.
func rotation(h int64) []*listing {
	ix := namedApps
	fresh := sort.Search(ix.Size(), func(i int) bool {
		_, v := ix.GetByIndex(i)
		return !v.(*listing).isNew(h)
	})
	old := ix.Size() - fresh
	if old == 0 {
		return nil
	}
	start := int((h / epochBlocks * rediscoverStep) % int64(old))
	out := window(ix, fresh+start, rediscoverSize)
	if wrap := rediscoverSize - len(out); wrap > 0 {
		if wrap > start {
			wrap = start // the window is wider than the old apps
		}
		out = append(out, window(ix, fresh, wrap)...)
	}
	return out
}

// topBuilders returns the namespaces with the most ranked stars, then the
// most listings, and the namespaces that started listing most recently, at
// most builderRank each, from indexes kept on every write.
func topBuilders() (top, fresh []*builder) {
	return builderRows(buildersByRank), builderRows(buildersByFirst)
}

func builderRows(index *avl.Tree) []*builder {
	var out []*builder
	index.Iterate("", "", func(_ string, v any) bool {
		out = append(out, v.(*builder))
		return len(out) >= builderRank
	})
	return out
}

// topN keeps the n best listings offered, by score descending then slug
// ascending, by insertion into a fixed-size slice.
type topN struct {
	n      int
	scores []int
	items  []*listing
}

func (t *topN) offer(score int, l *listing) {
	i := len(t.items)
	for i > 0 && (score > t.scores[i-1] || score == t.scores[i-1] && l.Slug < t.items[i-1].Slug) {
		i--
	}
	if i >= t.n {
		return
	}
	if len(t.items) < t.n {
		t.scores, t.items = append(t.scores, 0), append(t.items, nil)
	}
	copy(t.scores[i+1:], t.scores[i:])
	copy(t.items[i+1:], t.items[i:])
	t.scores[i], t.items[i] = score, l
}
