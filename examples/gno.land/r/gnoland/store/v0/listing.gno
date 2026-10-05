package store

import (
	"math"
	"strconv"
	"strings"

	"gno.land/p/nt/avl/v0"
)

type listing struct {
	Slug      string // unique among visible listings; Hide releases it
	Namespace string // of Path, for builders and the per-namespace cap
	Kind      string // app, service or package
	Path      string // the listing's identity
	Title     string // display name, unique among visible listings (titleKey)
	Tagline   string
	Category  string
	Palette   int
	Created   int64
	Updated   int64
	// Confirmed listings were listed by their own realm or by the seed, or
	// earned confirmStars ranked stars. Only they reach the shelves, the
	// feed and the builders: a claimed path is not verified, so a claim
	// alone buys no exposure (T11).
	Confirmed bool

	core        bool  // listed by the seed: shown in gno.land essentials, kept off New, Top and Rediscover
	live        rich  // rich fields currently in effect
	next        *rich // last submitted change, in effect from nextAt+richDelay
	nextAt      int64
	stars       *avl.Tree  // user address -> star, all of them (the displayed count)
	rankedStars *avl.Tree  // descKey(star height, address) -> star height, ranked stars only
	recent      *epochRing // ranked stars per epoch, for trending
	trendScore  int        // stars7d at the last ranked-star change, an upper bound of it since
	trendEpoch  int64      // epoch of that change
	milestone   int        // highest ranked-star milestone already announced
	hidden      bool       // hidden by GovDAO: kept under its path, shown nowhere
}

type star struct {
	Height int64
	// Ranked records whether the star counts in rankings (see Star).
	// Judged once, at star time, so reads need no cross-realm call.
	Ranked bool
}

type rich struct {
	Icon  string
	Cover string
}

// shelfIndex orders the visible, confirmed listings of one audience for the
// shelves.
type shelfIndex struct {
	created *avl.Tree // descKey(Created) -> *listing
	updated *avl.Tree // descKey(Updated) -> *listing, if Updated != Created
	stars   *avl.Tree // descKey(ranked stars) -> *listing, if any
}

// trendIndex files listings by the epoch of their last ranked-star change,
// one slot per epoch of the trending window, each ordered by trendScore.
// A listing's score is 0 once its slot is a window old, so a slot is simply
// dropped then: stale scores never pile up in front of live ones.
type trendIndex struct {
	slots [trendEpochs]trendSlot
}

type trendSlot struct {
	epoch   int64
	byScore *avl.Tree // descKey(trendScore) -> *listing
}

const trendEpochs = 7 // the trending window

func (t *trendIndex) set(l *listing) {
	s := &t.slots[l.trendEpoch%trendEpochs]
	switch {
	case l.trendScore < minTrending || l.trendEpoch < s.epoch: // nothing to rank, or stale
		return
	case s.byScore == nil || l.trendEpoch > s.epoch: // the slot's epoch left the window
		s.epoch, s.byScore = l.trendEpoch, avl.NewTree()
	}
	s.byScore.Set(descKey(int64(l.trendScore), l.Slug), l)
}

func (t *trendIndex) remove(l *listing) {
	if s := &t.slots[l.trendEpoch%trendEpochs]; s.byScore != nil && s.epoch == l.trendEpoch {
		s.byScore.Remove(descKey(int64(l.trendScore), l.Slug))
	}
}

// builder sums a namespace's visible, confirmed listings.
type builder struct {
	Namespace string
	Listings  int
	Stars     int   // their ranked stars
	First     int64 // creation height of the oldest
}

var (
	listings = avl.NewTree() // slug -> *listing, visible only
	byPath   = avl.NewTree() // path -> *listing, hidden included
	byTitle  = avl.NewTree() // titleKey -> slug of the visible listing using it
	byNS     = avl.NewTree() // namespace -> []*listing, hidden included

	// Visible listings only, maintained by index and unindex; the shelf
	// indexes and the builders hold confirmed ones only.
	byCategory = newCategoryIndex() // category -> *avl.Tree of descKey(Created) -> *listing, apps only
	apps       = newShelfIndex()
	toolbox    = newShelfIndex() // services and packages, for api/v1/build
	namedApps  = avl.NewTree()   // descKey(Created) -> *listing, named namespaces' apps but the core ones: New's and Rediscover's ground
	appTrend   = &trendIndex{}   // the build lens has no trending shelf

	// Builders, kept ranked on every write so reads never scan namespaces.
	builders        = avl.NewTree() // namespace -> *builder
	buildersByRank  = avl.NewTree() // descKey(rank, namespace) -> *builder
	buildersByFirst = avl.NewTree() // descKey(First, namespace) -> *builder
)

func newCategoryIndex() *avl.Tree {
	t := avl.NewTree()
	for _, c := range categories {
		t.Set(c.Key, avl.NewTree())
	}
	return t
}

func newShelfIndex() shelfIndex {
	return shelfIndex{avl.NewTree(), avl.NewTree(), avl.NewTree()}
}

// upsert lists l, or updates the listing of the same path in place, keeping
// its slug and creation height; a confirmed l confirms the listing. It
// returns a short reason when it refuses. It is the single write path for
// listing data, shared by Register, Claim and the genesis seed.
func upsert(l *listing, h int64) string {
	if paused {
		return "paused"
	}
	l.Namespace = namespaceOf(l.Path)
	existing := lookup(byPath, l.Path)
	if existing != nil {
		l.Slug = existing.Slug
	}
	if reason := validateListing(l); reason != "" {
		return reason
	}
	if existing != nil {
		if owner, _ := byTitle.Get(titleKey(l.Title)).(string); owner != "" && owner != l.Slug {
			return reasonTitleTaken
		}
		existing.update(l, h)
		if l.Confirmed {
			existing.confirm(h)
		}
		return ""
	}

	if reason := available(l); reason != "" {
		return reason
	}
	l.Created, l.Updated = h, h
	l.stars = avl.NewTree()
	l.rankedStars = avl.NewTree()
	l.recent = newEpochRing(trendEpochs)
	listings.Set(l.Slug, l)
	byPath.Set(l.Path, l)
	byTitle.Set(titleKey(l.Title), l.Slug)
	byNS.Set(l.Namespace, append(namespaceListings(l.Namespace), l))
	l.index()
	track("listed", l, h)
	return ""
}

// available returns why l, not visible, cannot take its slug, its title or
// a slot of its namespace, or "".
func available(l *listing) string {
	switch {
	case listings.Has(l.Slug):
		return reasonSlugTaken
	case byTitle.Has(titleKey(l.Title)):
		return reasonTitleTaken
	case visibleInNamespace(l.Namespace) >= maxPerNamespace:
		return reasonNamespaceFull
	}
	return ""
}

// update applies next's fields. Only a real change, at most once per
// updateCooldown, bumps Updated and is announced: re-registering in a loop
// must not buy a place in "Recently updated" (the fields still change, so a
// fix is never blocked). Only a kind or category change moves the listing
// in the indexes.
func (l *listing) update(next *listing, h int64) {
	if l.Kind == next.Kind && l.Title == next.Title && l.Tagline == next.Tagline &&
		l.Category == next.Category && l.Palette == next.Palette {
		return
	}
	if old, key := titleKey(l.Title), titleKey(next.Title); old != key {
		dropPick(l)    // GovDAO picked the app under its old title
		if !l.hidden { // a hidden listing holds no title
			byTitle.Remove(old)
			byTitle.Set(key, l.Slug)
		}
	}
	if l.Kind != next.Kind || l.Category != next.Category {
		l.reindex(func() { l.Kind, l.Category = next.Kind, next.Category })
	}
	l.Title, l.Tagline, l.Palette = next.Title, next.Tagline, next.Palette
	l.touch(h)
}

// touch records a real change at height h, unless one was already recorded
// less than updateCooldown ago.
func (l *listing) touch(h int64) {
	if h-l.Updated < updateCooldown {
		return
	}
	if l.indexed() {
		ix := l.shelves()
		ix.updated.Remove(descKey(l.Updated, l.Slug))
		ix.updated.Set(descKey(h, l.Slug), l)
	}
	l.Updated = h
	track("updated", l, h)
}

// confirm promotes a claimed listing to the shelves, the feed and the
// builders, and announces it.
func (l *listing) confirm(h int64) {
	if l.Confirmed {
		return
	}
	l.reindex(func() { l.Confirmed = true })
	track("listed", l, h)
}

// indexed reports whether l is on the shelves and counts for its builder.
func (l *listing) indexed() bool {
	return !l.hidden && l.Confirmed
}

// shownApp reports whether l is on the store front's shelves.
func (l *listing) shownApp() bool {
	return l.indexed() && l.Kind == kindApp
}

func (l *listing) shelves() shelfIndex {
	if l.Kind == kindApp {
		return apps
	}
	return toolbox
}

// index adds a visible app to its category, and a confirmed listing to the
// shelves and its builder. unindex undoes it; both run with l unchanged.
func (l *listing) index() {
	key := descKey(l.Created, l.Slug)
	if l.Kind == kindApp {
		categoryMembers(l.Category).Set(key, l)
	}
	if !l.Confirmed {
		return
	}
	ix := l.shelves()
	ix.created.Set(key, l)
	if l.Kind == kindApp && !address(l.Namespace).IsValid() && !l.core {
		namedApps.Set(key, l)
	}
	if l.Updated != l.Created {
		ix.updated.Set(descKey(l.Updated, l.Slug), l)
	}
	l.indexStars(ix)
	updateBuilder(l.Namespace, func(b *builder) {
		if b.Listings == 0 || l.Created < b.First {
			b.First = l.Created
		}
		b.Listings++
		b.Stars += l.rankedStars.Size()
	})
}

func (l *listing) unindex() {
	key := descKey(l.Created, l.Slug)
	if l.Kind == kindApp {
		categoryMembers(l.Category).Remove(key)
	}
	if !l.Confirmed {
		return
	}
	ix := l.shelves()
	ix.created.Remove(key)
	namedApps.Remove(key)
	ix.updated.Remove(descKey(l.Updated, l.Slug))
	l.unindexStars(ix)
	updateBuilder(l.Namespace, func(b *builder) {
		b.Listings--
		b.Stars -= l.rankedStars.Size()
		if b.First == l.Created {
			b.First = firstCreated(l.Namespace, l)
		}
	})
}

// indexStars and unindexStars keep l's entries in the star-ordered indexes.
func (l *listing) indexStars(ix shelfIndex) {
	if n := l.rankedStars.Size(); n > 0 {
		ix.stars.Set(descKey(int64(n), l.Slug), l)
		if l.Kind == kindApp {
			appTrend.set(l)
		}
	}
}

func (l *listing) unindexStars(ix shelfIndex) {
	if n := l.rankedStars.Size(); n > 0 {
		ix.stars.Remove(descKey(int64(n), l.Slug))
		if l.Kind == kindApp {
			appTrend.remove(l)
		}
	}
}

// updateBuilder applies change to ns's builder and moves it in the builder
// rankings whose key changed, dropping a builder left with no listing. The
// core namespaces have no builder: the rankings are the community's.
func updateBuilder(ns string, change func(b *builder)) {
	if isCoreNamespace(ns) {
		return
	}
	b, _ := builders.Get(ns).(*builder)
	fresh := b == nil
	if fresh {
		b = &builder{Namespace: ns}
		builders.Set(ns, b)
	}
	rank, first := b.rank(), b.First
	change(b)
	if b.Listings == 0 {
		builders.Remove(ns)
		buildersByRank.Remove(descKey(rank, ns))
		buildersByFirst.Remove(descKey(first, ns))
		return
	}
	rekey(buildersByRank, descKey(rank, ns), descKey(b.rank(), ns), b)
	if fresh || b.First != first { // never on a star
		rekey(buildersByFirst, descKey(first, ns), descKey(b.First, ns), b)
	}
}

// rank orders builders by ranked stars, then by listings: maxPerNamespace
// visible listings fit in the low 4 bits (TestBuilderRankFits).
func (b *builder) rank() int64 {
	return int64(b.Stars)*16 + int64(b.Listings)
}

// rekey moves v from key old to key next in tree, if they differ or v is
// not there yet.
func rekey(tree *avl.Tree, old, next string, v any) {
	if old == next && tree.Has(next) {
		return
	}
	tree.Remove(old)
	tree.Set(next, v)
}

// firstCreated is the creation height of the oldest visible, confirmed
// listing of ns other than except; a namespace has few listings.
func firstCreated(ns string, except *listing) int64 {
	first := int64(math.MaxInt64)
	for _, l := range namespaceListings(ns) {
		if l != except && l.indexed() && l.Created < first {
			first = l.Created
		}
	}
	return first
}

// reindex applies change to l and keeps the visible indexes in step. A
// listing that stops being a shown app loses GovDAO's pick for good: the
// vote was for the app as it was shown then.
func (l *listing) reindex(change func()) {
	if l.hidden {
		change()
		return
	}
	l.unindex()
	change()
	l.index()
	if !l.shownApp() {
		dropPick(l)
	}
}

// restar applies a change to l's ranked stars at height h and moves l in
// the star indexes and its builder's total only, so a star costs O(log n),
// not a full reindex.
func (l *listing) restar(h int64, change func()) {
	shown, ix, before := l.indexed(), l.shelves(), l.rankedStars.Size()
	if shown {
		l.unindexStars(ix)
	}
	change()
	l.trendScore, l.trendEpoch = l.stars7d(h), h/epochBlocks
	if !shown {
		return
	}
	l.indexStars(ix)
	updateBuilder(l.Namespace, func(b *builder) { b.Stars += l.rankedStars.Size() - before })
}

// addStar records a star. Only ranked stars rank a listing (Top,
// trending, milestones, builders) and confirm a claimed one; the others,
// from unnamed, newly named or owner addresses, only count in the displayed
// total.
func (l *listing) addStar(addr string, s star) {
	epoch := s.Height / epochBlocks
	l.stars.Set(addr, s)
	pulseStars.add(epoch, 1)
	if !s.Ranked {
		return
	}
	l.recent.add(epoch, 1)
	l.restar(s.Height, func() { l.rankedStars.Set(descKey(s.Height, addr), s.Height) })
	n := l.rankedStars.Size()
	if n >= confirmStars {
		l.confirm(s.Height)
	}
	for _, m := range milestones {
		if n == m && m > l.milestone {
			l.milestone = m
			track("milestone", l, s.Height)
		}
	}
}

// removeStar drops addr's star at height h. Once the star's epoch has left
// the rings, it no longer counts there, so nothing is subtracted.
func (l *listing) removeStar(addr string, h int64) {
	s := l.stars.Get(addr).(star)
	epoch := s.Height / epochBlocks
	l.stars.Remove(addr)
	pulseStars.add(epoch, -1)
	if !s.Ranked {
		return
	}
	l.recent.add(epoch, -1)
	l.restar(h, func() { l.rankedStars.Remove(descKey(s.Height, addr)) })
}

// descKey orders by n descending, then key ascending, in a forward
// iteration of an avl tree. descValue reads n back.
func descKey(n int64, key string) string {
	s := strconv.FormatInt(math.MaxInt64-n, 10)
	return strings.Repeat("0", 19-len(s)) + s + "/" + key
}

func descValue(key string) int64 {
	v, _ := strconv.ParseInt(key[:19], 10, 64)
	return math.MaxInt64 - v
}

// categoryMembers returns the visible apps of a validated category, newest
// first.
func categoryMembers(category string) *avl.Tree {
	return byCategory.Get(category).(*avl.Tree)
}

// appCount is the number of visible apps in a category.
func appCount(category string) int {
	return categoryMembers(category).Size()
}

func namespaceListings(ns string) []*listing {
	ls, _ := byNS.Get(ns).([]*listing)
	return ls
}

// visibleInNamespace counts the listings that use one of the namespace's
// maxPerNamespace slots: hidden ones release theirs. Hidden listings stay in
// byNS, so this grows only with GovDAO hides.
func visibleInNamespace(ns string) int {
	n := 0
	for _, l := range namespaceListings(ns) {
		if !l.hidden {
			n++
		}
	}
	return n
}

// titleKey makes title uniqueness blind to what validText cannot refuse
// (T6): case, spaces and separators, and the dashes and dots that look like
// them ("Gno-Land", "Gno‐Land", "Gno land", "gnoland"), and
// look-alike letters, folded to one Latin letter, so "GnoIand", "Gno1and",
// "Gnoiand" and a Cyrillic "АСМЕ" collide with "Gnoland" and "ACME".
func titleKey(title string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(title) {
		if r = foldRune(r); r != 0 {
			b.WriteRune(r)
		}
	}
	k := b.String()
	k = strings.ReplaceAll(k, "rn", "m")
	return strings.ReplaceAll(k, "vv", "w")
}

// foldRune maps a lower-case rune to the Latin letter it looks like, or to
// 0 for a separator titleKey drops.
func foldRune(r rune) rune {
	switch r {
	case ' ', '-', '_', '.', // separators, and the dashes and dots drawn like them
		'\u2010', '\u2011', '\u2012', '\u2013', '\u2014', '\u2015', '\u2043', '\u2212', '\uFE58', '\uFE63',
		'\u00B7', '\u2024', '\u2027', '\u2219', '\u22C5', '\u30FB':
		return 0
	case 'i', '1', '|', 'ı', 'і', 'ι':
		return 'l'
	case '0', 'о', 'ο':
		return 'o'
	case 'а', 'α':
		return 'a'
	case 'в', 'β':
		return 'b'
	case 'с':
		return 'c'
	case 'ԁ':
		return 'd'
	case 'е', 'ε':
		return 'e'
	case 'н', 'η':
		return 'h'
	case 'ј':
		return 'j'
	case 'к', 'κ':
		return 'k'
	case 'м':
		return 'm'
	case 'р', 'ρ':
		return 'p'
	case 'ԛ':
		return 'q'
	case 'ѕ':
		return 's'
	case 'т', 'τ':
		return 't'
	case 'υ':
		return 'u'
	case 'ν':
		return 'v'
	case 'ԝ':
		return 'w'
	case 'х', 'χ':
		return 'x'
	case 'у':
		return 'y'
	}
	return r
}

// lookup returns the listing stored under key, or nil.
func lookup(tree *avl.Tree, key string) *listing {
	l, _ := tree.Get(key).(*listing)
	return l
}

func bySlugOrPanic(slug string) *listing {
	l := lookup(listings, slug)
	if l == nil {
		panic(errUnknownListing)
	}
	return l
}

func byPathOrPanic(path string) *listing {
	l := lookup(byPath, path)
	if l == nil {
		panic("store: caller is not a listed realm")
	}
	return l
}

// effective returns the rich fields in effect at height h, and whether a
// newer change is still waiting for its publication delay.
func (l *listing) effective(h int64) (rich, bool) {
	switch {
	case l.next == nil:
		return l.live, false
	case h-l.nextAt >= richDelay:
		return *l.next, false
	default:
		return l.live, true
	}
}

// submitRich replaces any change still pending: the delay restarts.
// Resubmitting the latest values changes nothing.
func (l *listing) submitRich(r rich, h int64) {
	latest := l.live
	if l.next != nil {
		latest = *l.next
	}
	if r == latest {
		return
	}
	l.live, _ = l.effective(h)
	l.next = &r
	l.nextAt = h
	l.touch(h)
}

func (l *listing) isNew(h int64) bool {
	return h-l.Created < newWindow
}

func (l *listing) stars7d(h int64) int {
	return l.recent.sum(h/epochBlocks, trendEpochs)
}

// earned reports earned reputation (ADR §2) from chain data alone. Only
// ranked stars at least starMaturity old count (T8). Ranked stars are kept
// newest first, so the check reads one entry: the earnStars-th oldest must
// be starMaturity old.
func (l *listing) earned(h int64) bool {
	if h-l.Created < earnAge || l.rankedStars.Size() < earnStars {
		return false
	}
	if _, pending := l.effective(h); pending {
		return false
	}
	mature := false
	l.rankedStars.ReverseIterateByOffset(earnStars-1, 1, func(_ string, v any) bool {
		mature = h-v.(int64) >= starMaturity
		return true
	})
	return mature
}
