package store

import (
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
	blocksPerDay  = 17280

	gridSize = 8 // cards of a grid shelf
	rowSize  = 4 // cards of a full grid row on wide screens

	// minShelf hides a section too short to read as one; its apps are still
	// in their category and the lists.
	minShelf = 4

	// momentumMin is the weekly count worth showing: "+2 this week" reads
	// as a quiet chain, so smaller counts show nothing.
	momentumMin = 5
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
	Community bool   // neither the operator's nor established: says so on the path line
	Trust     string // "Core" or "Established", the positive badge of the other tiers
	Momentum  int    // ranked stars this week, when worth showing

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

// moment is the latest store event, closing the pulse line.
type moment struct {
	Text string
	URL  string
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
	Surprise  string // path of a random app
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

func (h *Handler) card(l *listing, labels map[string]string) card {
	c := card{
		Title:     l.Title,
		Tagline:   l.Tagline,
		KindLabel: kindLabels[l.Kind],
		Category:  category{Key: l.Category, Label: labels[l.Category]},
		Path:      l.webPath,
		ShortPath: l.shortPath,
		Stars:     l.Ranked,
		AllStars:  l.Stars,
		StarURL:   h.deps.RealmPath + "$help&func=Star&slug=" + l.Slug,
		Art:       l.art,
		Glyph:     l.glyph,
	}
	if l.tier.rich(l.Earned) {
		c.Trust = "Established"
		if l.tier == tierTrusted {
			c.Trust = "Core"
		}
		// Validated by sanitize: IPFS gateway URLs only.
		c.Icon = template.URL(l.Icon)   //nolint:gosec
		c.Cover = template.URL(l.Cover) //nolint:gosec
	} else {
		c.Community = true
	}
	if c.Category.Label == "" {
		c.Category.Label = l.Category
	}
	return c
}

func (h *Handler) cards(ls []*listing, labels map[string]string) []card {
	out := make([]card, len(ls))
	for i, l := range ls {
		out[i] = h.card(l, labels)
	}
	return out
}

// shelves resolves the realm's shelves into grids, leaving out the apps of
// skip (the Spotlight's). The realm decides order and rotation; gnoweb only
// enforces visual rules: a grid shows gridSize cards at most, in full rows
// of rowSize once it has one, and weekly momentum when it is worth it. A
// pinned shelf is drawn from its own head, with its empty text if it holds
// no app, unless all its apps are featured above or it has no empty text to
// show; the others show no app already shown above them and need minShelf
// cards.
func (h *Handler) shelves(src []shelf, bySlug map[string]*listing, labels map[string]string, skip map[string]bool) []shelfView {
	var out []shelfView
	shown := maps.Clone(skip)
	if shown == nil {
		shown = make(map[string]bool)
	}
	for i, s := range src {
		left := shown
		if s.Pinned {
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
		week := make(map[string]int, len(s.Stars7d))
		for k, n := range s.Stars7d {
			week[s.Slugs[k]] = n
		}
		v := shelfView{ID: "store-shelf-" + strconv.Itoa(i), Title: s.Title, Note: s.Note, Empty: s.Empty, Cards: h.cards(picked, labels), More: s.More}
		for k, l := range picked {
			shown[l.Slug] = true
			if n := week[l.Slug]; n >= momentumMin {
				v.Cards[k].Momentum = n
			}
		}
		out = append(out, v)
	}
	return out
}

// eligible returns the first spotlightPool apps allowed to look official
// (trusted or earned), in the realm's shelf order, which puts quality first,
// and how many of them earned their place rather than being the operator's.
func eligible(src []shelf, bySlug map[string]*listing) (pool []*listing, earned int) {
	used := make(map[string]bool)
	for _, s := range src {
		for _, slug := range s.Slugs {
			l, ok := bySlug[slug]
			if !ok || used[slug] || l.Kind != kindApp || !l.tier.rich(l.Earned) {
				continue
			}
			used[slug] = true
			if l.tier != tierTrusted {
				earned++
			}
			if pool = append(pool, l); len(pool) == spotlightPool {
				return pool, earned
			}
		}
	}
	return pool, earned
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
// with the chain's height, so every gnoweb shows the same and caches it.
func featured(res *homeResponse, bySlug map[string]*listing) (hero *listing, label string, spot []*listing) {
	pool, earned := eligible(res.Shelves, bySlug)
	day := int(res.Height / blocksPerDay)
	if pick := bySlug[res.Pick]; pick != nil && pick.Kind == kindApp {
		hero, label = pick, labelPick
	} else if l := trending(res.Shelves, bySlug, pool); l != nil {
		hero, label = l, labelPopular
		if leads(res.Shelves, l.Slug) {
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
	capped := earned >= spotlightSize
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

// trending returns the eligible app with the most stars this week, if that
// is worth showing (momentumMin).
func trending(src []shelf, bySlug map[string]*listing, pool []*listing) (best *listing) {
	most := momentumMin - 1
	for _, s := range src {
		for i, n := range s.Stars7d {
			if l := bySlug[s.Slugs[i]]; l != nil && n > most && slices.Contains(pool, l) {
				best, most = l, n
			}
		}
	}
	return best
}

// leads reports whether slug is first, ties included, on the realm's
// trending shelf (the one behind the "trending" list), counting every app
// on it, eligible or not: only then may the hero say "Trending now".
func leads(src []shelf, slug string) bool {
	for _, s := range src {
		if !slices.ContainsFunc(s.More, func(l listRef) bool { return l.Key == "trending" }) {
			continue
		}
		i := slices.Index(s.Slugs, slug)
		return i >= 0 && len(s.Stars7d) > 0 && s.Stars7d[i] == slices.Max(s.Stars7d)
	}
	return false
}

func (h *Handler) buildHome(res *homeResponse) homeData {
	labels := categoryLabels(res.Categories)
	bySlug := indexListings(res.Listings)
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
		Pulse:     newPulse(res.chainClock, res.Pulse, latestMoment(res.Activity, bySlug)),
		Spotlight: h.cards(spot, labels),
		Shelves:   h.shelves(res.Shelves, bySlug, labels, skip),
	}
	if hero != nil {
		data.Hero = &heroView{card: h.card(hero, labels), Label: label}
		if _, _, ok := offered(res, pagedList, "trending"); ok && (label == labelTrending || label == labelPopular) {
			data.Hero.LabelList = "trending"
		}
	}
	if n := len(res.Listings); n > 0 {
		data.Surprise = res.Listings[rand.IntN(n)].webPath //nolint:gosec // a random app to visit, not a secret
	}
	return data
}

// buildBuild needs no category labels: services and packages show their
// kind, not a category.
// Empty builder lists are left out, so the section is too when both are.
func (h *Handler) buildBuild(res *buildResponse) buildData {
	data := buildData{Intro: buildLine, Shelves: h.shelves(res.Shelves, indexListings(res.Listings), nil, nil)}
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

// latestMoment phrases the newest store event it can with fixed templates;
// only titles come from listings, and only listings that passed sanitize.
func latestMoment(events []activity, bySlug map[string]*listing) *moment {
	for _, e := range events {
		l, ok := bySlug[e.Slug]
		if !ok {
			continue
		}
		var text string
		switch e.Kind {
		case "listed":
			text = l.Title + " just joined"
		case "updated":
			text = l.Title + " was updated"
		case "milestone":
			if e.Stars <= 0 {
				continue
			}
			text = l.Title + " crossed " + strconv.Itoa(e.Stars) + " stars"
		default:
			continue
		}
		return &moment{Text: text, URL: l.webPath}
	}
	return nil
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
// is cut in its middle ("g17zyd…9cxg"), its two ends kept: a look-alike
// namespace must never be hidden by truncation (ADR-004 T6). The full path
// is in the element's title.
func shortPath(pkgPath string) string {
	ns, rest, _ := strings.Cut(pkgPath, "/")
	if isAddress(ns) {
		ns = components.TruncMiddle(ns, 6, 4)
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
