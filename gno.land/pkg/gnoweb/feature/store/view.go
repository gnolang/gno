package store

import (
	"cmp"
	"context"
	"html/template"
	"maps"
	"math/rand/v2"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
)

// kindLabels name what is not an app; an app is named by its category.
var kindLabels = map[string]string{kindService: "Service", kindPackage: "Package"}

const (
	// spotlightSize apps of the spotlightPool best eligible ones are shown,
	// a new window each day.
	spotlightSize = 3
	spotlightPool = 12
	spotlightMin  = 2 // below this, the Spotlight is left out
	secondsPerDay = 24 * 60 * 60

	gridSize = 8 // cards of a grid shelf
	rowSize  = 4 // cards of a full grid row on wide screens

	// minShelf hides a section too short to read as one; its apps are still
	// in their category and the lists.
	minShelf = 4

	// momentumMin is the weekly count worth showing: "+2 this week" reads
	// as a quiet chain, so smaller counts show nothing.
	momentumMin = 5

	// momentAge bounds how old the latest event may be to close the pulse:
	// on an idle store, an app is not "just joined" for months.
	momentAge = 7 * 24 * time.Hour

	// usedMin is the distinct callers over the week worth showing, on a card
	// and for Most used: fewer read as a quiet chain, as with momentumMin.
	usedMin = 3
)

// card is one listing as every template draws it. It links to the
// listing's own realm or package page: the store has no detail page of its
// own, so nothing is shown twice.
type card struct {
	Title     string
	Tagline   string
	KindLabel string // "Package" or "Service"; empty for an app, which shows its category
	Category  category
	Path      string // gnoweb path: the card's link
	ShortPath string // the path as shortPath shortens it, the full one in its title
	Stars     int    // ranked stars, the ones that order the store
	AllStars  int    // every star
	StarURL   string // the store's Star action, prefilled for this listing
	Community bool   // neither the operator's nor established: says so on the path line, its links not followed
	// Unverified marks a listing claimed by its namespace's owner that
	// neither its own realm nor ranked stars confirm. It says so in place
	// of "community".
	Unverified bool
	Trust      string // "Core" (the realm's core apps) or "Established" (earned), a positive badge
	Momentum   int    // ranked stars this week, when worth showing
	Users      int    // distinct callers this week, from an indexer, when worth showing

	// Owner assets, set only when the tier allows them.
	Icon  template.URL
	Cover template.URL

	// Generated visuals, always set: no card is ever blank.
	Art   template.HTML
	Glyph template.HTML
}

// shelfView is a shelf, drawn as a grid of cards.
type shelfView struct {
	ID    string // of its heading, which labels the section
	Title string
	Note  string
	Empty string // shown in place of the grid when there is no card
	Cards []card
	More  []listRef // the paged lists behind it
}

// catNav is the category strip on top of the app pages. On a list page
// and on a page that does not exist, no category is current, not even All.
type catNav struct {
	Items     []category
	Active    string
	NoCurrent bool
}

// listTab is one entry of the list switcher on list pages.
type listTab struct {
	Key     string
	Label   string
	Current bool
}

// pulseItem is one figure of the pulse, with the window it counts.
type pulseItem struct {
	Value  int
	Label  string
	Window string
}

type pulseView struct {
	Height int64
	Ago    string
	Items  []pulseItem
	Moment *moment // the latest store event
}

// appLink links to a listing's own page; Community as on card.
type appLink struct {
	URL       string
	Community bool
}

// moment is the latest store event, closing the pulse line.
type moment struct {
	Text string
	appLink
}

// heroView is the page's one hero, with why it is there. LabelList, if
// set, is the list the label links to.
type heroView struct {
	card
	Label     string
	LabelList string
}

type homeData struct {
	Intro     string // the lead, also the head description
	Nav       catNav
	Surprise  *appLink // a random app
	MovedTo   string   // the successor store's web path, if GovDAO named one
	Pulse     pulseView
	Hero      *heroView
	Spotlight []card
	Shelves   []shelfView
}

type builderView struct {
	Name     string
	URL      string
	Listings int
	Stars    int
}

type builderSection struct {
	Title  string
	People []builderView
}

type buildData struct {
	Intro    string // the lead, also the head description
	Shelves  []shelfView
	Builders []builderSection
}

// pageData is one page of a category or a list. Missing marks a category,
// list or page that does not exist; Empty is shown in place of the grid.
type pageData struct {
	Nav     catNav
	Lists   []listTab // on list pages only
	Title   string
	Note    string
	Cards   []card
	Empty   string
	Missing bool
	Page    int
	Pages   int
	Prev    int // 0: no previous page
	Next    int // 0: no next page
}

// usage is the indexer's count of each realm's callers this week, at hand;
// nil without an indexer or while it has none. It never waits: a page is
// drawn without the figures rather than late.
func (h *Handler) usage(ctx context.Context) map[string]int {
	if h.deps.Usage == nil {
		return nil
	}
	return h.deps.Usage(ctx)
}

func (h *Handler) card(l *listing, labels map[string]string, usage map[string]int) card {
	c := card{
		Title:      l.Title,
		Tagline:    l.Tagline,
		KindLabel:  kindLabels[l.Kind],
		Category:   category{Key: l.Category, Label: labels[l.Category]},
		Path:       l.webPath,
		ShortPath:  l.shortPath,
		Stars:      l.Ranked,
		AllStars:   l.Stars,
		StarURL:    h.deps.RealmPath + "$help&func=Star&slug=" + l.Slug,
		Art:        l.art,
		Glyph:      l.glyph,
		Unverified: !l.Confirmed,
	}
	if n := usage[l.Path]; n >= usedMin {
		c.Users = n
	}
	// Core is the realm's word for its seeded apps, never the operator's
	// trust list, which only lets owner assets show.
	switch {
	case l.Core:
		c.Trust = "Core"
	case l.tier == tierRegistered && l.earned:
		c.Trust = "Established"
	}
	if c.Community = l.community(); !c.Community {
		// Validated by sanitize: IPFS gateway URLs only.
		c.Icon = template.URL(l.Icon)   //nolint:gosec
		c.Cover = template.URL(l.Cover) //nolint:gosec
	}
	if c.Category.Label == "" {
		c.Category.Label = l.Category
	}
	return c
}

func (h *Handler) cards(ls []*listing, labels map[string]string, usage map[string]int) []card {
	out := make([]card, len(ls))
	for i, l := range ls {
		out[i] = h.card(l, labels, usage)
	}
	return out
}

// shelves resolves the realm's shelves into grids, leaving out the apps of
// skip (the Spotlight's). The realm decides order and rotation; gnoweb words
// them, sets their floors (applyFloors) and enforces visual rules: a grid
// shows gridSize cards at most, in full rows of rowSize once it has one,
// and weekly momentum when it is worth it. A
// pinned shelf is drawn from its own head, with its empty text if it holds
// no app, unless all its apps are featured above or it has no empty text to
// show; the others show no app already shown above them and need minShelf
// cards.
func (h *Handler) shelves(src []shelf, bySlug map[string]*listing, labels map[string]string, skip map[string]bool, usage map[string]int) []shelfView {
	var out []shelfView
	shown := maps.Clone(skip)
	if shown == nil {
		shown = make(map[string]bool)
	}
	for i, s := range src {
		left := shown
		if s.Pinned || s.ownHead {
			left = skip
		}
		picked := pick(bySlug, s.Slugs, left)
		if len(picked) > gridSize {
			picked = picked[:gridSize]
		}
		// Full rows only: a lone card under a full row reads as a gap.
		// The trimmed apps are not marked shown, so a shelf below may
		// still show them.
		if len(picked) >= rowSize {
			picked = picked[:len(picked)/rowSize*rowSize]
		}
		if s.Pinned && len(picked) == 0 && (len(s.Slugs) > 0 || s.Empty == "") || !s.Pinned && len(picked) < minShelf {
			continue
		}
		v := shelfView{ID: "store-shelf-" + strconv.Itoa(i), Title: s.Title, Note: s.Note, Empty: s.Empty, Cards: h.cards(picked, labels, usage), More: s.More}
		for k, l := range picked {
			shown[l.Slug] = true
			if s.minWeek > 0 && l.Week >= momentumMin { // on a shelf ranked by the week
				v.Cards[k].Momentum = l.Week
			}
		}
		out = append(out, v)
	}
	return out
}

// eligiblePool returns up to spotlightPool eligible apps in the realm's
// shelf order, and whether spotlightSize of all eligible apps, wherever
// they sit, earned their place rather than being the operator's (capped).
// Once capped, the pool keeps the operator's first app only, so a shelf of
// them ahead of the earned ones cannot fill it.
func eligiblePool(src []shelf, bySlug map[string]*listing) (pool []*listing, capped bool) {
	var all []*listing
	used := make(map[string]bool)
	earned := 0
	for _, s := range src {
		for _, slug := range s.Slugs {
			l, ok := bySlug[slug]
			if !ok || used[slug] || !l.eligible() {
				continue
			}
			used[slug] = true
			if l.tier != tierTrusted {
				earned++
			}
			all = append(all, l)
		}
	}
	capped = earned >= spotlightSize
	trusted := 0
	for _, l := range all {
		if capped && l.tier == tierTrusted {
			if trusted++; trusted > 1 {
				continue
			}
		}
		if pool = append(pool, l); len(pool) == spotlightPool {
			break
		}
	}
	return pool, capped
}

// Why the hero is there.
const (
	labelPick     = "Picked by GovDAO" // GovDAO's vote is the endorsement
	labelTrending = "Trending now"     // the store's #1 this week
	labelPopular  = "Popular this week"
	labelRotation = "In the spotlight today"
)

// featured chooses the page's hero and its Spotlight, with no editor: a
// GovDAO pick first; else the eligible app with the most momentum this week,
// called trending only if it leads the realm's trending list outright;
// else a daily rotation over the eligible apps. The Spotlight rotates daily
// through the rest. Only eligible apps are shown, so a community app cannot
// take the most visible place on the page by stars alone. Once
// spotlightSize apps have earned their place, the operator's own apps take
// one place at most, hero included: the store must not look like it
// promotes its host; until then, at launch, they fill it. Everything moves
// with the chain's block time, so every gnoweb shows the same and caches it.
func featured(res *homeResponse, bySlug map[string]*listing) (hero *listing, label string, spot []*listing) {
	pool, capped := eligiblePool(res.Shelves, bySlug)
	day := int(res.Time / secondsPerDay)
	if pick := bySlug[res.Pick]; pick != nil && pick.Kind == kindApp {
		hero, label = pick, labelPick
	} else if l := trending(res.Shelves, bySlug); l != nil {
		hero, label = l, labelPopular
		if leads(res.Shelves, bySlug, l) {
			label = labelTrending
		}
	} else if len(pool) > 0 {
		hero, label = pool[day%len(pool)], labelRotation
	} else {
		return nil, "", nil
	}

	// One ordered candidate list, the hero then the day's window of the
	// pool, kept in one pass: each app once, and the operator's capped.
	candidates := make([]*listing, 0, len(pool)+1)
	candidates = append(candidates, hero)
	for i := range pool {
		candidates = append(candidates, pool[(day*spotlightSize+i)%len(pool)])
	}
	trusted := 0
	kept := make([]*listing, 0, 1+spotlightSize)
	for _, l := range candidates {
		if slices.Contains(kept, l) || capped && trusted > 0 && l.tier == tierTrusted {
			continue
		}
		if l.tier == tierTrusted {
			trusted++
		}
		if kept = append(kept, l); len(kept) > spotlightSize {
			break
		}
	}
	if spot = kept[1:]; len(spot) < spotlightMin {
		spot = nil
	}
	return hero, label, spot
}

// trendingKey is the realm's shelf of the apps most starred this week.
const trendingKey = "trending"

// trending returns the eligible app of the realm's trending shelf with the
// most stars this week, if that is worth showing (momentumMin).
func trending(src []shelf, bySlug map[string]*listing) (best *listing) {
	most := momentumMin - 1
	for _, s := range src {
		if s.Key != trendingKey {
			continue
		}
		for _, slug := range s.Slugs {
			if l := bySlug[slug]; l != nil && l.Week > most && l.eligible() {
				best, most = l, l.Week
			}
		}
	}
	return best
}

// leads reports whether l is first, ties included, on the realm's trending
// shelf, counting every app on it, eligible or not: only then may the hero
// say "Trending now".
func leads(src []shelf, bySlug map[string]*listing, l *listing) bool {
	for _, s := range src {
		if s.Key != trendingKey || !slices.Contains(s.Slugs, l.Slug) {
			continue
		}
		return !slices.ContainsFunc(s.Slugs, func(slug string) bool {
			other := bySlug[slug]
			return other != nil && other.Week > l.Week
		})
	}
	return false
}

// mostUsed is the shelf of the front page's apps people call most, by
// distinct callers this week, when an indexer says so. It ranks the apps the
// front page already carries: no read is added for it. A caller costs one
// funded address, less than a ranked star, so the shelf takes no pinned
// place (withMostUsed) and, like New, leaves free g1… namespaces out.
func mostUsed(res *homeResponse, usage map[string]int) (shelf, bool) {
	var used []*listing
	for i := range res.Listings {
		if l := &res.Listings[i]; l.Kind == kindApp && l.tier != tierAnonymous && usage[l.Path] >= usedMin {
			used = append(used, l)
		}
	}
	slices.SortFunc(used, func(a, b *listing) int {
		return cmp.Or(cmp.Compare(usage[b.Path], usage[a.Path]), cmp.Compare(a.Slug, b.Slug))
	})
	s := shelf{Key: "used", shelfDef: shelfDef{
		Title:   "Most used this week",
		Note:    "By distinct callers over the last 7 days, as this site's indexer counts them.",
		ownHead: true,
	}}
	for _, l := range used {
		s.Slugs = append(s.Slugs, l.Slug)
	}
	return s, len(used) > 0
}

// withMostUsed returns the front page's shelves with Most used, if any,
// first of those that are not pinned: after New and Top, which keep the
// head of the page whatever an indexer counts. The slice is a new one: res
// is shared.
func withMostUsed(res *homeResponse, usage map[string]int) []shelf {
	used, ok := mostUsed(res, usage)
	if !ok {
		return res.Shelves
	}
	at := slices.IndexFunc(res.Shelves, func(s shelf) bool { return !s.Pinned })
	if at < 0 {
		at = len(res.Shelves)
	}
	return slices.Insert(slices.Clone(res.Shelves), at, used)
}

func (h *Handler) buildHome(res *homeResponse, usage map[string]int) homeData {
	labels := categoryLabels(res.Categories)
	bySlug := indexListings(res.Listings)
	shelves := withMostUsed(res, usage)
	hero, label, spot := featured(res, bySlug)
	skip := make(map[string]bool, len(spot)+1)
	if hero != nil {
		skip[hero.Slug] = true
	}
	for _, l := range spot {
		skip[l.Slug] = true
	}
	data := homeData{
		Intro:     introLine,
		Nav:       catNav{Items: res.Categories},
		Pulse:     newPulse(res.chainClock, res.Pulse, latestMoment(res.Shelves, bySlug, time.Now())),
		MovedTo:   res.MovedTo,
		Spotlight: h.cards(spot, labels, usage),
		Shelves:   h.shelves(shelves, bySlug, labels, skip, usage),
	}
	if hero != nil {
		data.Hero = &heroView{card: h.card(hero, labels, usage), Label: label}
		if _, _, _, ok := offered(res, pagedList, trendingKey); ok && (label == labelTrending || label == labelPopular) {
			data.Hero.LabelList = trendingKey
		}
	}
	if n := len(res.Listings); n > 0 {
		surprise := res.Listings[rand.IntN(n)].link() //nolint:gosec // a random app to visit, not a secret
		data.Surprise = &surprise
	}
	return data
}

// buildBuild needs no category labels: services and packages show their
// kind, not a category.
// Empty builder lists are left out, so the section is too when both are.
func (h *Handler) buildBuild(res *buildResponse, usage map[string]int) buildData {
	data := buildData{Intro: buildLine, Shelves: h.shelves(res.Shelves, indexListings(res.Listings), nil, nil, usage)}
	for _, s := range []builderSection{
		{"Top builders", builders(res.Builders.Top)},
		{"New builders", builders(res.Builders.New)},
	} {
		if len(s.People) > 0 {
			data.Builders = append(data.Builders, s)
		}
	}
	return data
}

// newPulse keeps, for each figure, the shortest window that is not zero: a
// quiet chain reads as "this month", never as "0 this week", and a figure
// with nothing to show is left out rather than padded. A window is said
// once for the figures in a row that share it. The realm counts listings
// of every kind (listed) and every star given, net of those taken back.
func newPulse(clock chainClock, p pulse, m *moment) pulseView {
	v := pulseView{Height: clock.Height, Moment: m}
	if clock.Time > 0 {
		v.Ago = components.FormatRelativeTimeSince(time.Unix(clock.Time, 0))
	}
	last := ""
	for _, f := range []struct {
		one, many string
		week, mon int
	}{
		{"app or package listed", "apps & packages listed", p.Listed7d, p.Listed30d},
		{"star given", "stars given", p.Stars7d, p.Stars30d},
	} {
		n, window := f.week, "this week"
		if n <= 0 {
			n, window = f.mon, "this month"
		}
		if n <= 0 {
			continue
		}
		it := pulseItem{Value: n, Label: f.many}
		if n == 1 {
			it.Label = f.one
		}
		if window != last {
			it.Window, last = window, window
		}
		v.Items = append(v.Items, it)
	}
	return v
}

// latestMoment tells the newest thing the front page's own listings show:
// an app listed or changed within momentAge, among those of the realm's New
// and Recently updated shelves, which hold named namespaces' confirmed apps.
// The store keeps no feed of events. An app that is still community is
// named by its path, not its title: the top of the page is no place for
// words anyone can register (ADR-004 T6).
func latestMoment(shelves []shelf, bySlug map[string]*listing, now time.Time) *moment {
	var (
		latest *listing
		at     int64
		what   string
	)
	for _, s := range shelves {
		for _, slug := range s.Slugs {
			l := bySlug[slug]
			if l == nil {
				continue
			}
			switch {
			case s.Key == "new" && l.Created > at:
				latest, at, what = l, l.Created, " just joined"
			case s.Key == "updated" && l.Updated > at:
				latest, at, what = l, l.Updated, " was updated"
			}
		}
	}
	if latest == nil || now.Sub(time.Unix(at, 0)) > momentAge {
		return nil
	}
	name := latest.Title
	if latest.community() {
		name = latest.shortPath
	}
	return &moment{Text: name + what, appLink: latest.link()}
}

// link is the way to a listing's own page.
func (l *listing) link() appLink {
	return appLink{URL: l.webPath, Community: l.community()}
}

func builders(bs []builder) []builderView {
	out := make([]builderView, len(bs))
	for i, b := range bs {
		out[i] = builderView{
			Name:     b.Namespace,
			URL:      "/u/" + b.Namespace,
			Listings: b.Listings,
			Stars:    b.Stars,
		}
	}
	return out
}

// shortPath keeps the namespace first and elides the middle of deep paths
// only, down to the package's name, with its version if it has one:
// "acme/apps/v2/swap" → "acme/…/swap", "acme/x/swap/v2" → "acme/…/swap/v2",
// while "gnoland/boards/v0" stays as it is. A namespace that is an address
// is cut in its middle ("g17zyd8dhz…ex9cxg"), its two ends kept: a
// look-alike namespace must never be hidden by truncation (ADR-004 T6). The
// 8 data and 6 checksum characters kept are 70 bits, out of reach of a
// vanity key search. The full path is in the element's title.
func shortPath(pkgPath string) string {
	ns, rest, _ := strings.Cut(pkgPath, "/")
	if isAddress(ns) {
		ns = components.TruncMiddle(ns, 10, 6)
	}
	if strings.Count(rest, "/") < 2 {
		return ns + "/" + rest
	}
	return ns + "/…/" + components.PackageName(pkgPath)
}

func categoryLabels(cs []category) map[string]string {
	m := make(map[string]string, len(cs))
	for _, c := range cs {
		m[c.Key] = c.Label
	}
	return m
}

func ptrs(ls []listing) []*listing {
	out := make([]*listing, len(ls))
	for i := range ls {
		out[i] = &ls[i]
	}
	return out
}

func indexListings(ls []listing) map[string]*listing {
	m := make(map[string]*listing, len(ls))
	for i := range ls {
		m[ls[i].Slug] = &ls[i]
	}
	return m
}

// pick resolves slugs against the referenced listings, once each, skipping
// unknown ones and those of exclude.
func pick(bySlug map[string]*listing, slugs []string, exclude map[string]bool) []*listing {
	out := make([]*listing, 0, len(slugs))
	inList := make(map[string]bool, len(slugs))
	for _, slug := range slugs {
		l, ok := bySlug[slug]
		if !ok || inList[slug] || exclude[slug] {
			continue
		}
		inList[slug] = true
		out = append(out, l)
	}
	return out
}
