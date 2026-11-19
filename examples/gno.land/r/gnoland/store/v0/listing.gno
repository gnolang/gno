package store

import (
	"chain"
	"chain/runtime"
	"math"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"gno.land/p/nt/avl/v0"
)

type listing struct {
	Slug      string // unique among visible listings; Hide releases it, and a claim's gives way to a realm listing itself (reslugClaim)
	Namespace string // of Path, for builders and the per-namespace cap
	Kind      string // app, service or package
	Path      string // the listing's identity
	Title     string // display name, unique among visible confirmed listings (titleKey)
	Tagline   string
	Category  string
	Palette   int
	Created   int64
	Updated   int64
	// Confirmed listings were listed by their own realm or by the seed, or
	// earned confirmStars ranked stars. Only they reach the shelves and the
	// builders, and only they hold their title: a claimed path is not
	// verified, so a claim alone buys no exposure and reserves no name (T11).
	Confirmed bool
	// selfListed listings were listed by their own realm or the seed: their
	// confirmation stands for good. A claim confirmed by stars keeps it only
	// while it has confirmStars ranked stars.
	selfListed bool
	// announced listings are counted in the pulse: a claim unconfirmed and
	// confirmed again is not a new listing, and must not count in a loop.
	announced bool

	core       bool   // listed by the seed: shown in gno.land essentials, kept off New, Top and Rediscover
	icon       string // rich fields currently in effect, kept flat: a rich value is one more stored object
	cover      string
	next       *rich // last submitted change, in effect from nextAt+richDelay
	nextAt     int64
	book       *starBook // nil until the first star, so an unstarred listing stores no star objects
	trendScore int       // stars7d at the last ranked-star change, an upper bound of it since
	trendEpoch int64     // epoch of that change
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

const trendEpochs = 7 // the trending window

// trendKey files an app in appTrend by the epoch of its last ranked-star
// change, newest first, then by its trendScore.
func trendKey(l *listing) string {
	return descKey(l.trendEpoch, descKey(int64(l.trendScore), l.Slug))
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
	byTitle  avl.Tree // titleKey -> slug of the visible confirmed listing holding it

	// hiddenCore counts the core listings GovDAO has hidden: they keep their
	// slug and their title (slugTaken, titleTaken).
	hiddenCore int

	// Visible listings only, maintained by index and unindex; the shelf
	// indexes and the builders hold confirmed ones only.
	byCategory   avl.Tree // category -> *avl.Tree of catKey -> *listing, apps only; made by a category's first app
	apps         shelfIndex
	toolbox      shelfIndex // services and packages, for api/v1/build
	namedApps    avl.Tree   // descKey(Created) -> *listing, named namespaces' apps but the core ones: New's and Rediscover's ground
	namedUpdated avl.Tree   // descKey(Updated) -> *listing, named namespaces' apps: Recently updated's ground
	// trendKey -> *listing, shown apps with a trendScore: an app has one
	// entry at most, so no write drops entries in bulk, and one older than
	// the trending window is just not read. The build lens does not trend.
	appTrend avl.Tree

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
	existing, asked := lookup(&byPath, l.Path), l.Slug
	if existing != nil {
		if existing.core {
			return reasonCore // frozen in code (ADR §5), for its realm and its namespace's owner alike
		}
		l.Slug = existing.Slug
		// What the listing already says, from who already listed it: a
		// Relist, or an upgraded realm's init. Nothing to check again.
		if existing.says(l) && (existing.selfListed || !l.Confirmed) {
			existing.confirmByStars(now) // its title may be free by now
			existing.takeSlug(asked)
			return ""
		}
	}
	if reason := validateListing(l); reason != "" {
		return reason
	}
	if existing != nil {
		// A title is checked when it changes hands: asked for, or about to
		// be held by a claim its realm confirms.
		if (l.Title != existing.Title || l.Confirmed && !existing.Confirmed) && titleTaken(l) {
			return reasonTitleTaken
		}
		existing.update(l, now)
		if l.Confirmed {
			existing.selfListed = true
			existing.confirm(now)
		}
		existing.confirmByStars(now) // under a title now free, the stars it has count
		existing.takeSlug(asked)
		return ""
	}

	if !validSlug(l.Slug) {
		return "bad_slug"
	}
	if reason := available(l); reason != "" {
		return reason
	}
	reslugClaim(l.Slug)
	l.selfListed = l.Confirmed
	l.Created, l.Updated = now, now
	listings.Set(l.Slug, l)
	byPath.Set(l.Path, l)
	l.index()
	l.announce(now)
	return ""
}

// titleHolder returns the slug of the visible confirmed listing holding l's
// title, or "".
func titleHolder(l *listing) string {
	slug, _ := byTitle.Get(titleKey(l.Title)).(string)
	return slug
}

// titleTaken reports whether a listing other than l's own holds l's title:
// a visible confirmed one, or a hidden core one, which keeps its name
// (hidePath). A claim holds none, so a title is never taken from anyone.
func titleTaken(l *listing) bool {
	key := titleKey(l.Title)
	slug, _ := byTitle.Get(key).(string)
	// By path, not by slug: a hidden listing shares its slug with the
	// listing that took it.
	if holder := lookup(&listings, slug); holder != nil && holder.Path != l.Path {
		return true
	}
	if hiddenCore == 0 {
		return false
	}
	for _, c := range essentials {
		if c.hidden && c.Path != l.Path && titleKey(c.Title) == key {
			return true
		}
	}
	return false
}

// slugTaken reports whether l cannot have slug. A visible listing holds its
// own, but a claim gives way to a confirmed l: a claim proves nothing (T11),
// and an app whose init is refused may never call Register again. A hidden
// core listing keeps its own too.
func slugTaken(l *listing, slug string) bool {
	if holder := lookup(&listings, slug); holder != nil && (holder.Confirmed || !l.Confirmed) {
		return true
	}
	if hiddenCore == 0 {
		return false
	}
	for _, c := range essentials {
		if c.hidden && c != l && c.Slug == slug {
			return true
		}
	}
	return false
}

// reslugClaim moves the unconfirmed claim under slug, if any, to a slug of
// its own (movedSlug), once slugTaken has let a confirmed listing have it.
// The claim keeps its path, its stars and its place: nothing its owner or
// its fans paid to store is deleted for another listing's sake, so there is
// no deposit to collect by listing over it.
func reslugClaim(slug string) {
	c := lookup(&listings, slug)
	if c == nil || c.Confirmed {
		return
	}
	c.reslug(movedSlug(slug))
	chain.Emit("StoreClaimMoved", "path", c.Path, "from", slug, "slug", c.Slug)
}

// reslug moves the visible listing l to slug, which is free.
func (l *listing) reslug(slug string) {
	l.unindex()
	listings.Remove(l.Slug)
	l.Slug = slug
	listings.Set(slug, l)
	l.index()
}

// takeSlug gives a visible listing whose slug the store moved (reslugClaim)
// the slug it asks for, once it may have it as a new listing would: the
// listing that took its own may be hidden by now, and a package, which is
// only ever a claim, would otherwise never get its slug back. Any other
// listing keeps its slug for good, and so does this one from then on.
func (l *listing) takeSlug(slug string) {
	if slug == l.Slug || l.hidden || !strings.Contains(l.Slug, movedMark) || !validSlug(slug) || slugTaken(l, slug) {
		return
	}
	reslugClaim(slug)
	l.reslug(slug)
}

// moved numbers the slugs given to moved claims.
var moved int

// movedSlug returns the slug a listing moves to from slug: slug, cut to
// fit, then movedMark and a number never given before. No listing can ask
// for a slug holding movedMark (validSlug), so it is free by construction,
// with no search, and is never taken from the listing: it leaves it by
// asking for another (takeSlug).
func movedSlug(slug string) string {
	moved++
	suffix := movedMark + strconv.Itoa(moved)
	if len(slug)+len(suffix) > maxSlug {
		slug = slug[:maxSlug-len(suffix)]
	}
	return slug + suffix
}

// available returns why l, a new listing, cannot take its slug, its title
// or a slot of its namespace, or "". reslugClaim then frees a slug an
// unconfirmed claim holds.
func available(l *listing) string {
	if slugTaken(l, l.Slug) {
		return reasonSlugTaken
	}
	return roomFor(l)
}

// roomFor returns why l, not visible, cannot hold its title or a slot of
// its namespace, or "".
func roomFor(l *listing) string {
	switch {
	case titleTaken(l) && (l.Confirmed || !l.hidden):
		// A claim held no title before its hide, so a title taken since
		// keeps no unhide from it: confirmByStars checks it again.
		return reasonTitleTaken
	case !l.core && !isCoreNamespace(l.Namespace) && visibleInNamespace(l.Namespace) >= maxPerNamespace:
		// The core namespaces hold the seed, and room for what follows it;
		// a core listing of another namespace is never kept from an unhide.
		return reasonNamespaceFull
	}
	return ""
}

// says reports whether l already says what next does.
func (l *listing) says(next *listing) bool {
	return l.Kind == next.Kind && l.Title == next.Title && l.Tagline == next.Tagline &&
		l.Category == next.Category && l.Palette == next.Palette
}

// update applies next's fields. Only a real change, at most once per
// updateCooldown, bumps Updated: re-registering in a loop must not buy a
// place in "Recently updated" (the fields still change, so a fix is never
// blocked). Only a change of kind, category or title moves the listing in
// the indexes.
func (l *listing) update(next *listing, now int64) {
	if l.says(next) {
		return
	}
	retitled := l.Title != next.Title && titleKey(l.Title) != titleKey(next.Title)
	if retitled {
		dropPick(l) // GovDAO picked the app under its old title
	}
	if retitled || l.Kind != next.Kind || l.Category != next.Category {
		l.reindex(func() { l.Kind, l.Category, l.Title = next.Kind, next.Category, next.Title })
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
// namespace earns no front-page slot (New, Recently updated, Rediscover).
func (l *listing) namedApp() bool {
	return l.Kind == kindApp && !address(l.Namespace).IsValid()
}

// confirm promotes a claimed listing to the shelves and the builders, where
// it holds its title (index). While another listing holds that title it
// stays a claim: its owner claims it again under another.
func (l *listing) confirm(now int64) {
	// A hidden listing holds no title: unhide checks it.
	if l.Confirmed || !l.hidden && titleTaken(l) {
		return
	}
	l.reindex(func() { l.Confirmed = true })
	l.announce(now)
}

// confirmByStars confirms a visible claim that has the ranked stars for it.
// A hidden one waits for its unhide, which checks its title: confirmed
// while hidden, it would need a title it never held to come back.
func (l *listing) confirmByStars(now int64) {
	if !l.hidden && l.rankedCount() >= confirmStars {
		l.confirm(now)
	}
}

// announce counts l among the pulse's new listings, once.
func (l *listing) announce(now int64) {
	if !l.announced && l.indexed() {
		l.announced = true
		pulseListed.add(now/epochLength, 1)
	}
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

// catKey files an app in its category: confirmed apps first, each group
// newest first. A claim costs a free address little, so claims must never
// push a confirmed app down its page.
func (l *listing) catKey() string {
	group := "1"
	if l.Confirmed {
		group = "0"
	}
	return group + descKey(l.Created, l.Slug)
}

// index adds a visible app to its category, and a confirmed listing to the
// shelves and its builder, where it holds its title. unindex undoes it;
// both run with l unchanged.
func (l *listing) index() {
	key := descKey(l.Created, l.Slug)
	if l.Kind == kindApp {
		members, ok := byCategory.Get(l.Category).(*avl.Tree)
		if !ok {
			members = avl.NewTree()
			byCategory.Set(l.Category, members)
		}
		members.Set(l.catKey(), l)
	}
	if !l.Confirmed {
		return
	}
	byTitle.Set(titleKey(l.Title), l.Slug)
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
		categoryMembers(l.Category).Remove(l.catKey())
	}
	if !l.Confirmed {
		return
	}
	byTitle.Remove(titleKey(l.Title))
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
		if l.Kind == kindApp && l.trendScore >= minTrending {
			appTrend.Set(trendKey(l), l)
		}
	}
}

func (l *listing) unindexStars(ix *shelfIndex) {
	if n := l.rankedCount(); n > 0 {
		ix.stars.Remove(descKey(int64(n), l.Slug))
		if l.Kind == kindApp {
			appTrend.Remove(trendKey(l))
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
// trending, builders) and confirm a claimed one; the others,
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
	l.confirmByStars(s.At)
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
		l.reindex(func() { l.Confirmed = false }) // its backers left: a claim again, which holds no title
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

// appCount is the number of visible apps in a category, claims included:
// the number its pages hold.
func appCount(category string) int {
	return categoryMembers(category).Size()
}

// confirmedCount is the number of confirmed apps in a category. They are
// filed first (catKey), so it is the rank of the first claim.
func confirmedCount(category string) int {
	members := categoryMembers(category)
	return sort.Search(members.Size(), func(i int) bool {
		key, _ := members.GetByIndex(i)
		return key[0] == '1'
	})
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
		// Most titles are plain ASCII, for which the general path below is
		// lower-casing and foldRune's first cases: the same key, without
		// walking its switch for every letter (TestTitleKeyASCII).
		if r < utf8.RuneSelf {
			switch {
			case r >= 'A' && r <= 'Z':
				r += 'a' - 'A'
			case r == ' ' || r == '-' || r == '_' || r == '.' || r == 0:
				continue
			case r == '1' || r == '|':
				r = 'l'
			case r == '0':
				r = 'o'
			case r == '`':
				r = '\''
			case r == '"':
				b.WriteString("''")
				continue
			}
			if r == 'i' {
				r = 'l'
			}
			b.WriteByte(byte(r))
			continue
		}
		switch {
		case unicode.IsMark(r): // a mark only redraws its letter
			continue
		case r == 'Æ' || r == 'æ':
			b.WriteString("ae")
			continue
		case r == 'Œ' || r == 'œ':
			b.WriteString("oe")
			continue
		case r == 'Ĳ' || r == 'ĳ':
			b.WriteString("lj")
			continue
		case r == 'ŉ':
			b.WriteString("'n")
			continue
		case r == 'Ы' || r == 'ы': // drawn as ЬІ
			b.WriteString("bl")
			continue
		case r == 'Њ' || r == 'њ': // НЬ
			b.WriteString("hb")
			continue
		case r == 'Љ' || r == 'љ': // ЛЬ
			b.WriteString("лb")
			continue
		case r == 'แ': // Thai sara ae, drawn as two sara e
			b.WriteString("เเ")
			continue
		case r == 'װ': // Hebrew double vav, double yod, vav yod
			b.WriteString("ll")
			continue
		case r == 'ײ':
			b.WriteString("''")
			continue
		case r == 'ױ':
			b.WriteString("l'")
			continue
		}
		if c := foldCapital(r); c != 0 {
			r = c
		} else {
			r = unicode.ToLower(r)
		}
		if base := latinBase(r); base != 0 {
			r = base
		}
		switch r = foldRune(r); r {
		case 0:
		case '"':
			b.WriteString("''")
		default:
			b.WriteRune(r)
		}
	}
	k := b.String()
	k = strings.ReplaceAll(k, "rn", "m")
	return strings.ReplaceAll(k, "vv", "w")
}

// foldCapital maps the Greek capitals drawn as Latin ones whose lower case
// foldRune does not fold to that letter, or returns 0: Nu is an N, its
// lower case a v. The other Greek and Cyrillic capitals lower-case to a
// letter foldRune folds.
func foldCapital(r rune) rune {
	switch r {
	case 'Ά':
		return 'a'
	case 'Έ':
		return 'e'
	case 'Η', 'Ή': // an H, its lower case an n
		return 'h'
	case 'Ί':
		return 'l'
	case 'Μ':
		return 'm'
	case 'Ν':
		return 'n'
	case 'Ό':
		return 'o'
	case 'Υ', 'Ύ', 'Ϋ':
		return 'y'
	case 'Ζ':
		return 'z'
	case 'Γ': // Greek capitals that only draw as Cyrillic ones
		return 'г'
	case 'Π':
		return 'п'
	case 'Φ':
		return 'ф'
	}
	return 0
}

// foldRune maps a lower-case rune to the Latin letter it looks like, or to
// 0 for a separator titleKey drops.
func foldRune(r rune) rune {
	switch r {
	case ' ', '-', '_', '.', // separators, and the dashes and dots drawn like them
		'\u2010', '\u2011', '\u2012', '\u2013', '\u2014', '\u2015', '\u2043', '\u2212', '\uFE58', '\uFE63',
		'\u00B7', '\u2024', '\u2027', '\u2219', '\u22C5', '\u30FB',
		'ㅡ', '一', prolonged: // Hangul eu, Han one and the prolonged sound mark, drawn as a dash
		return 0
	case 'i', '1', '|', 'ı', 'і', 'ї', 'ι', 'ί', 'ϊ', 'ΐ',
		'ا', 'أ', 'إ', 'آ', 'ٱ', 'ٲ', 'ٵ', 'ו', 'ן', 'ㅣ', '丨': // Arabic alef, Hebrew vav and final nun, Hangul i, Han stroke
		return 'l'
	case '0', 'о', 'ο', 'ό', 'ס', 'ه', 'ە', 'ہ', 'ھ', 'ㅇ': // Hebrew samekh, Arabic heh, Hangul ieung
		return 'o'
	case '3', 'з': // Cyrillic ze
		return '3'
	case '6', 'б': // Cyrillic be
		return '6'
	case 'а', 'α', 'ά':
		return 'a'
	case 'в', 'β', 'ь', 'ъ':
		return 'b'
	case 'с':
		return 'c'
	case 'е', 'ё', 'ѐ', 'ε', 'έ':
		return 'e'
	case 'н', 'ћ':
		return 'h'
	case 'ј':
		return 'j'
	case 'к', 'ќ', 'κ':
		return 'k'
	case 'м':
		return 'm'
	case 'η', 'ή', 'п', 'π':
		return 'n'
	case 'р', 'ρ':
		return 'p'
	case 'г', 'ґ', 'ѓ':
		return 'r'
	case 'ѕ', 'ऽ': // Devanagari avagraha
		return 's'
	case 'т', 'τ', '丁', '丅':
		return 't'
	case 'υ', 'ύ', 'ϋ', 'ΰ', 'μ':
		return 'u'
	case 'ν':
		return 'v'
	case 'ω', 'ώ':
		return 'w'
	case 'х', 'χ':
		return 'x'
	case 'у', 'ў', 'γ', '丫':
		return 'y'
	case '乙':
		return 'z'
	case '十':
		return '+'
	case 'ノ', '丿':
		return '/'
	case 'ॽ': // Devanagari glottal stop
		return '?'
	case '‘', '’', '`', 'י': // typographic quotes, the grave accent and Hebrew yod, drawn as an apostrophe
		return '\''
	case '“', '”': // and as a double quote, which titleKey writes as two apostrophes
		return '"'
	case '\u0958': // Devanagari letters with a nukta, as the letter and the nukta titleKey drops
		return 'क'
	case '\u0959':
		return 'ख'
	case '\u095A':
		return 'ग'
	case '\u095B':
		return 'ज'
	case '\u095C':
		return 'ड'
	case '\u095D':
		return 'ढ'
	case '\u095E':
		return 'फ'
	case '\u095F':
		return 'य'
	case 'ک': // Persian keheh and yeh, and alef maksura, drawn as Arabic kaf and yeh
		return 'ك'
	case 'ی', 'ى':
		return 'ي'
	case 'φ': // Greek and Cyrillic letters drawn alike
		return 'ф'
	case '力': // Han and hiragana drawn as katakana
		return 'カ'
	case '口', '囗':
		return 'ロ'
	case '\uFA11': // a variant of 崎 (山﨑, 山崎)
		return '崎'
	case '匕':
		return 'ヒ'
	case '亻':
		return 'イ'
	case '厶':
		return 'ム'
	case '工':
		return 'エ'
	case '二':
		return 'ニ'
	case '八':
		return 'ハ'
	case '夕':
		return 'タ'
	case '卜':
		return 'ト'
	case 'へ':
		return 'ヘ'
	case 'べ':
		return 'ベ'
	case 'ぺ':
		return 'ペ'
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
