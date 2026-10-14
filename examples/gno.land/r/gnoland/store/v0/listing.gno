package store

import (
	"chain"
	"chain/runtime"
	"math"
	"strconv"
	"strings"
	"unicode"

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
	// selfListed listings were listed by their own realm or the seed: their
	// confirmation stands for good. A claim confirmed by stars keeps it only
	// while it has confirmStars ranked stars.
	selfListed bool

	core       bool   // listed by the seed: shown in gno.land essentials, kept off New, Top and Rediscover
	icon       string // rich fields currently in effect, kept flat: a rich value is one more stored object
	cover      string
	next       *rich // last submitted change, in effect from nextAt+richDelay
	nextAt     int64
	book       *starBook // nil until the first star, so an unstarred listing stores no star objects
	trendScore int       // stars7d at the last ranked-star change, an upper bound of it since
	trendEpoch int64     // epoch of that change
	milestone  int       // highest ranked-star milestone already announced
	hidden     bool      // hidden by GovDAO: kept under its path, shown nowhere
}

// starBook holds a listing's stars. Its trees are values, not pointers, so
// a first star stores one object fewer per tree.
type starBook struct {
	all    avl.Tree  // user address -> packed star, all of them (the displayed count)
	ranked avl.Tree  // descKey(star time, address) -> star time, ranked stars only
	recent epochRing // ranked stars per epoch, for trending; empty until the first ranked star
}

type star struct {
	At int64 // block time of the star
	// Ranked records whether the star counts in rankings (see Star).
	// Judged once, at star time, so reads need no cross-realm call.
	Ranked bool
}

// pack stores s as one int64, its time doubled plus its rank bit: a
// struct value would be one more stored object per star.
func (s star) pack() int64 {
	v := s.At * 2
	if s.Ranked {
		v++
	}
	return v
}

// starOf returns addr's star on l, which must exist.
func (l *listing) starOf(addr string) star {
	v := l.book.all.Get(addr).(int64)
	return star{At: v / 2, Ranked: v%2 == 1}
}

// starCount is the displayed number of stars; rankedCount, the number that
// rank the listing.
func (l *listing) starCount() int {
	if l.book == nil {
		return 0
	}
	return l.book.all.Size()
}

func (l *listing) rankedCount() int {
	if l.book == nil {
		return 0
	}
	return l.book.ranked.Size()
}

// starred reports whether addr stars l.
func (l *listing) starred(addr string) bool {
	return l.book != nil && l.book.all.Has(addr)
}

type rich struct {
	Icon  string
	Cover string
}

// shelfIndex orders the visible, confirmed listings of one audience for the
// shelves.
type shelfIndex struct {
	created avl.Tree // descKey(Created) -> *listing
	updated avl.Tree // descKey(Updated) -> *listing, if Updated != Created
	stars   avl.Tree // descKey(ranked stars) -> *listing, if any
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
	First     int64 // creation time of the oldest
}

var (
	listings avl.Tree // slug -> *listing, visible only
	byPath   avl.Tree // path -> *listing, hidden included; also read per namespace (namespaceListings)
	byTitle  avl.Tree // titleKey -> slug of the visible listing using it

	// Visible listings only, maintained by index and unindex; the shelf
	// indexes and the builders hold confirmed ones only.
	byCategory   avl.Tree // category -> *avl.Tree of descKey(Created) -> *listing, apps only; made by a category's first app
	apps         shelfIndex
	toolbox      shelfIndex // services and packages, for api/v1/build
	namedApps    avl.Tree   // descKey(Created) -> *listing, named namespaces' apps but the core ones: New's and Rediscover's ground
	namedUpdated avl.Tree   // descKey(Updated) -> *listing, named namespaces' apps: Recently updated's ground
	appTrend     trendIndex // the build lens has no trending shelf

	// Builders, kept ranked on every write so reads never scan namespaces.
	builders        avl.Tree // namespace -> *builder
	buildersByRank  avl.Tree // descKey(rank, namespace) -> *builder
	buildersByFirst avl.Tree // descKey(First, namespace) -> *builder
)

// upsert lists l, or updates the listing of the same path in place, keeping
// its slug and creation time; a confirmed l confirms the listing. It
// returns a short reason when it refuses. It is the single write path for
// listing data, shared by Register, Claim and the genesis seed.
func upsert(l *listing, now int64) string {
	if paused {
		return "paused"
	}
	l.Namespace = namespaceOf(l.Path)
	existing := lookup(&byPath, l.Path)
	if existing != nil {
		l.Slug = existing.Slug
	}
	if reason := validateListing(l); reason != "" {
		return reason
	}
	if existing != nil {
		if owner := titleHolder(l); owner != "" && owner != l.Slug {
			if !l.Confirmed || !evictClaim(owner) {
				return reasonTitleTaken
			}
		}
		existing.update(l, now)
		if l.Confirmed {
			existing.selfListed = true
			existing.confirm(now)
		}
		return ""
	}

	if l.Confirmed { // an unverified claim of l's slug or title gives way
		evictClaim(l.Slug)
		evictClaim(titleHolder(l))
	}
	if reason := available(l); reason != "" {
		return reason
	}
	l.selfListed = l.Confirmed
	l.Created, l.Updated = now, now
	listings.Set(l.Slug, l)
	byPath.Set(l.Path, l)
	byTitle.Set(titleKey(l.Title), l.Slug)
	l.index()
	track("listed", l, now)
	return ""
}

// titleHolder returns the slug of the visible listing using l's title, or "".
func titleHolder(l *listing) string {
	slug, _ := byTitle.Get(titleKey(l.Title)).(string)
	return slug
}

// evictClaim drops the listing under slug if it is an unconfirmed claim,
// and reports whether it did. It runs when an app's own Register or the seed
// wants its slug or title: a claim proves nothing (T11), and an app whose
// init is refused never calls Register again, so the claim would keep it out
// for good.
func evictClaim(slug string) bool {
	c := lookup(&listings, slug)
	if c == nil || c.Confirmed {
		return false
	}
	c.unindex()
	listings.Remove(c.Slug)
	byTitle.Remove(titleKey(c.Title))
	byPath.Remove(c.Path)
	chain.Emit("StoreClaimDropped", "slug", c.Slug, "path", c.Path)
	return true
}

// available returns why l, not visible, cannot take its slug, its title or
// a slot of its namespace, or "".
func available(l *listing) string {
	switch {
	case listings.Has(l.Slug):
		return reasonSlugTaken
	case byTitle.Has(titleKey(l.Title)):
		return reasonTitleTaken
	case !isCoreNamespace(l.Namespace) && visibleInNamespace(l.Namespace) >= maxPerNamespace:
		return reasonNamespaceFull // the core namespaces hold the seed, and room for what follows it
	}
	return ""
}

// update applies next's fields. Only a real change, at most once per
// updateCooldown, bumps Updated and is announced: re-registering in a loop
// must not buy a place in "Recently updated" (the fields still change, so a
// fix is never blocked). Only a kind or category change moves the listing
// in the indexes.
func (l *listing) update(next *listing, now int64) {
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
	l.touch(now)
}

// touch records a real change at time now, unless one was already recorded
// less than updateCooldown ago.
func (l *listing) touch(now int64) {
	if now-l.Updated < updateCooldown {
		return
	}
	shown := l.indexed()
	if shown {
		l.removeUpdated(l.shelves())
	}
	l.Updated = now
	if shown {
		l.setUpdated(l.shelves())
	}
	track("updated", l, now)
}

// setUpdated and removeUpdated keep l's entries in the indexes ordered by
// Updated.
func (l *listing) setUpdated(ix *shelfIndex) {
	key := descKey(l.Updated, l.Slug)
	ix.updated.Set(key, l)
	if l.namedApp() {
		namedUpdated.Set(key, l)
	}
}

func (l *listing) removeUpdated(ix *shelfIndex) {
	key := descKey(l.Updated, l.Slug)
	ix.updated.Remove(key)
	namedUpdated.Remove(key)
}

// namedApp reports an app under a named namespace: a free g1… address
// namespace earns no front-page slot (New, Recently updated, Rediscover) and
// no place in the activity feed.
func (l *listing) namedApp() bool {
	return l.Kind == kindApp && !address(l.Namespace).IsValid()
}

// confirm promotes a claimed listing to the shelves, the feed and the
// builders, and announces it.
func (l *listing) confirm(now int64) {
	if l.Confirmed {
		return
	}
	l.reindex(func() { l.Confirmed = true })
	track("listed", l, now)
}

// indexed reports whether l is on the shelves and counts for its builder.
func (l *listing) indexed() bool {
	return !l.hidden && l.Confirmed
}

// shownApp reports whether l is on the store front's shelves.
func (l *listing) shownApp() bool {
	return l.indexed() && l.Kind == kindApp
}

func (l *listing) shelves() *shelfIndex {
	if l.Kind == kindApp {
		return &apps
	}
	return &toolbox
}

// index adds a visible app to its category, and a confirmed listing to the
// shelves and its builder. unindex undoes it; both run with l unchanged.
func (l *listing) index() {
	key := descKey(l.Created, l.Slug)
	if l.Kind == kindApp {
		members, ok := byCategory.Get(l.Category).(*avl.Tree)
		if !ok {
			members = avl.NewTree()
			byCategory.Set(l.Category, members)
		}
		members.Set(key, l)
	}
	if !l.Confirmed {
		return
	}
	ix := l.shelves()
	ix.created.Set(key, l)
	if l.namedApp() && !l.core {
		namedApps.Set(key, l)
	}
	if l.Updated != l.Created {
		l.setUpdated(ix)
	}
	l.indexStars(ix)
	updateBuilder(l.Namespace, func(b *builder) {
		if b.Listings == 0 || l.Created < b.First {
			b.First = l.Created
		}
		b.Listings++
		b.Stars += l.rankedCount()
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
	l.removeUpdated(ix)
	l.unindexStars(ix)
	updateBuilder(l.Namespace, func(b *builder) {
		b.Listings--
		b.Stars -= l.rankedCount()
		if b.First == l.Created {
			b.First = firstCreated(l.Namespace, l)
		}
	})
}

// indexStars and unindexStars keep l's entries in the star-ordered indexes.
func (l *listing) indexStars(ix *shelfIndex) {
	if n := l.rankedCount(); n > 0 {
		ix.stars.Set(descKey(int64(n), l.Slug), l)
		if l.Kind == kindApp {
			appTrend.set(l)
		}
	}
}

func (l *listing) unindexStars(ix *shelfIndex) {
	if n := l.rankedCount(); n > 0 {
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
	rekey(&buildersByRank, descKey(rank, ns), descKey(b.rank(), ns), b)
	if fresh || b.First != first { // never on a star
		rekey(&buildersByFirst, descKey(first, ns), descKey(b.First, ns), b)
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

// firstCreated is the creation time of the oldest visible, confirmed
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

// restar applies a change to l's ranked stars at time now and moves l in
// the star indexes and its builder's total only, so a star costs O(log n),
// not a full reindex.
func (l *listing) restar(now int64, change func()) {
	shown, ix, before := l.indexed(), l.shelves(), l.rankedCount()
	if shown {
		l.unindexStars(ix)
	}
	change()
	l.trendScore, l.trendEpoch = l.stars7d(now), now/epochLength
	if !shown {
		return
	}
	l.indexStars(ix)
	updateBuilder(l.Namespace, func(b *builder) { b.Stars += l.rankedCount() - before })
}

// addStar records a star. Only ranked stars rank a listing (Top,
// trending, milestones, builders) and confirm a claimed one; the others,
// from unnamed, newly named or owner addresses, only count in the displayed
// total.
func (l *listing) addStar(addr string, s star) {
	epoch := s.At / epochLength
	if l.book == nil {
		l.book = &starBook{}
	}
	l.book.all.Set(addr, s.pack())
	pulseStars.add(epoch, 1)
	if !s.Ranked {
		return
	}
	if l.book.recent.counts == nil {
		l.book.recent = newEpochRing(trendEpochs)
	}
	l.book.recent.add(epoch, 1)
	l.restar(s.At, func() { l.book.ranked.Set(descKey(s.At, addr), s.At) })
	n := l.rankedCount()
	if n >= confirmStars {
		l.confirm(s.At)
	}
	for _, m := range milestones {
		if n == m && m > l.milestone {
			l.milestone = m
			track("milestone", l, s.At)
		}
	}
}

// removeStar drops addr's star at time now. Once the star's epoch has left
// the rings, it no longer counts there, so nothing is subtracted.
func (l *listing) removeStar(addr string, now int64) {
	s := l.starOf(addr)
	epoch := s.At / epochLength
	l.book.all.Remove(addr)
	pulseStars.add(epoch, -1)
	if !s.Ranked {
		return
	}
	l.book.recent.add(epoch, -1)
	l.restar(now, func() { l.book.ranked.Remove(descKey(s.At, addr)) })
	if l.Confirmed && !l.selfListed && l.rankedCount() < confirmStars {
		l.reindex(func() { l.Confirmed = false }) // its backers left: a claim again
	}
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
// first. A category no app has used has no tree yet: it reads as a new,
// empty one, which nothing stores.
func categoryMembers(category string) *avl.Tree {
	if members, ok := byCategory.Get(category).(*avl.Tree); ok {
		return members
	}
	return avl.NewTree()
}

// appCount is the number of visible apps in a category.
func appCount(category string) int {
	return categoryMembers(category).Size()
}

// namespaceListings returns ns's listings, hidden included, from the
// byPath ranges of its realms and packages: a separate index would cost
// storage for what byPath already orders.
func namespaceListings(ns string) []*listing {
	var ls []*listing
	for _, kind := range []string{"/p/", "/r/"} {
		prefix := runtime.ChainDomain() + kind + ns + "/"
		// "0" follows "/": the range holds exactly the paths under prefix.
		byPath.Iterate(prefix, prefix[:len(prefix)-1]+"0", func(_ string, v any) bool {
			ls = append(ls, v.(*listing))
			return false
		})
	}
	return ls
}

// visibleInNamespace counts the listings that use one of the namespace's
// maxPerNamespace slots: hidden ones release theirs. Hidden listings stay in
// byPath, so this grows only with GovDAO hides.
func visibleInNamespace(ns string) int {
	n := 0
	for _, l := range namespaceListings(ns) {
		if !l.hidden {
			n++
		}
	}
	return n
}

// titleKey makes title uniqueness blind to what validTitle cannot refuse
// (T6): case, spaces and separators, and the dashes and dots that look like
// them ("Gno-Land", "Gno‐Land", "Gno land", "gnoland"), accents ("Bóards"),
// and look-alike letters, folded to one Latin letter, so "GnoIand",
// "Gno1and", "Gnoiand" and a Cyrillic "АСМЕ" or Greek "ΤΟΚΕΝ" collide with
// "Gnoland", "ACME" and "TOKEN". Capitals are folded before lower-casing:
// a Greek capital Nu draws as an N, its lower case as a v.
func titleKey(title string) string {
	var b strings.Builder
	for _, r := range title {
		if c := foldCapital(r); c != 0 {
			r = c
		} else {
			r = unicode.ToLower(r)
		}
		if base := latinBase(r); base != 0 {
			r = base
		}
		if r = foldRune(r); r != 0 {
			b.WriteRune(r)
		}
	}
	k := b.String()
	k = strings.ReplaceAll(k, "rn", "m")
	return strings.ReplaceAll(k, "vv", "w")
}

// foldCapital maps a Greek or Cyrillic capital drawn as a Latin capital to
// that Latin letter, lower-case, or returns 0.
func foldCapital(r rune) rune {
	switch r {
	case 'Α', 'Ά', 'А':
		return 'a'
	case 'Β', 'В':
		return 'b'
	case 'С':
		return 'c'
	case 'Ε', 'Έ', 'Е':
		return 'e'
	case 'Η', 'Ή', 'Н', 'Һ':
		return 'h'
	case 'Ι', 'Ί', 'І', 'Ӏ':
		return 'l'
	case 'Ј':
		return 'j'
	case 'Κ', 'К':
		return 'k'
	case 'Μ', 'М':
		return 'm'
	case 'Ν':
		return 'n'
	case 'Ο', 'Ό', 'О':
		return 'o'
	case 'Ρ', 'Р':
		return 'p'
	case 'Ԛ':
		return 'q'
	case 'Ѕ':
		return 's'
	case 'Τ', 'Т':
		return 't'
	case 'Ѵ':
		return 'v'
	case 'Ԝ':
		return 'w'
	case 'Χ', 'Х':
		return 'x'
	case 'Υ', 'Ύ', 'У', 'Ү':
		return 'y'
	case 'Ζ':
		return 'z'
	}
	return 0
}

// foldRune maps a lower-case rune to the Latin letter it looks like, or to
// 0 for a separator titleKey drops.
func foldRune(r rune) rune {
	switch r {
	case ' ', '-', '_', '.', // separators, and the dashes and dots drawn like them
		'\u2010', '\u2011', '\u2012', '\u2013', '\u2014', '\u2015', '\u2043', '\u2212', '\uFE58', '\uFE63',
		'\u00B7', '\u2024', '\u2027', '\u2219', '\u22C5', '\u30FB':
		return 0
	case 'i', '1', '|', 'ı', 'і', 'ι', 'ӏ', 'ا', 'ו', 'ן': // Arabic alef, Hebrew vav and final nun
		return 'l'
	case '0', 'о', 'ο', 'ס': // Hebrew samekh
		return 'o'
	case 'а', 'α':
		return 'a'
	case 'в', 'β':
		return 'b'
	case 'с', 'ϲ':
		return 'c'
	case 'ԁ':
		return 'd'
	case 'е', 'ε':
		return 'e'
	case 'н', 'η', 'һ':
		return 'h'
	case 'ј', 'ϳ':
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
	case 'ν', 'ѵ':
		return 'v'
	case 'ԝ':
		return 'w'
	case 'х', 'χ':
		return 'x'
	case 'у', 'ү', 'γ':
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
	l := lookup(&listings, slug)
	if l == nil {
		panic(errUnknownListing)
	}
	return l
}

func byPathOrPanic(path string) *listing {
	l := lookup(&byPath, path)
	if l == nil {
		panic("store: caller is not a listed realm")
	}
	return l
}

// live returns the rich fields in effect before any pending change.
func (l *listing) live() rich {
	return rich{Icon: l.icon, Cover: l.cover}
}

// effective returns the rich fields in effect at time now, and whether a
// newer change is still waiting for its publication delay.
func (l *listing) effective(now int64) (rich, bool) {
	switch {
	case l.next == nil:
		return l.live(), false
	case now-l.nextAt >= richDelay:
		return *l.next, false
	default:
		return l.live(), true
	}
}

// submitRich replaces any change still pending: the delay restarts.
// Resubmitting the latest values changes nothing.
func (l *listing) submitRich(r rich, now int64) {
	latest := l.live()
	if l.next != nil {
		latest = *l.next
	}
	if r == latest {
		return
	}
	live, _ := l.effective(now)
	l.icon, l.cover = live.Icon, live.Cover
	l.next = &r
	l.nextAt = now
	l.touch(now)
}

func (l *listing) isNew(now int64) bool {
	return now-l.Created < newWindow
}

func (l *listing) stars7d(now int64) int {
	if l.book == nil {
		return 0
	}
	return l.book.recent.sum(now/epochLength, trendEpochs)
}

// fifthRanked returns when l got its earnStars-th oldest ranked star, or 0
// with fewer: the one entry earned reputation needs, as ranked stars are
// kept newest first.
func (l *listing) fifthRanked() int64 {
	if l.rankedCount() < earnStars {
		return 0
	}
	var at int64
	l.book.ranked.ReverseIterateByOffset(earnStars-1, 1, func(_ string, v any) bool {
		at = v.(int64)
		return true
	})
	return at
}
