package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"time"
)

const (
	// apiVersion is the api/v1 contract this package reads. Unknown fields
	// are ignored; a different version is refused rather than misread.
	apiVersion = 1

	// fetchTimeout bounds one realm query, independent of its caller.
	fetchTimeout = 4 * time.Second

	// maxPayload refuses an oversized answer before decoding it.
	maxPayload = 2 << 20
)

// The paged endpoints: a category, or a list the front page opens onto.
const (
	pagedCategory = "category"
	pagedList     = "list"
)

var (
	errNotFound = errors.New("store: not found")
	errTooLarge = errors.New("store: response too large")
)

// envelope is embedded in every response.
type envelope struct {
	Version int    `json:"version"`
	Error   string `json:"error"`
}

func (e *envelope) check() error {
	switch {
	case e.Error == "not_found":
		return errNotFound
	case e.Error != "":
		return fmt.Errorf("store: realm error %q", e.Error)
	case e.Version != apiVersion:
		return fmt.Errorf("store: unsupported api version %d", e.Version)
	}
	return nil
}

// category is one of the realm's categories. Count is every visible app in
// it, the number its pages hold; Confirmed, those listed by their own realm
// or confirmed since, the number it is shown with. The realm files the
// others, claims no one has confirmed, after them.
type category struct {
	Key       string `json:"key"`
	Label     string `json:"label"`
	Count     int    `json:"count"`
	Confirmed int    `json:"confirmed"`
}

// shelf is a section of the front page or the build lens, in the realm's
// order. The realm names it by its key and fills it; how it reads and what
// it asks of an app are gnoweb's (shelfDef).
type shelf struct {
	Key   string    `json:"key"`
	Slugs []string  `json:"slugs"`
	More  []listRef `json:"more"` // the paged lists behind the shelf

	shelfDef `json:"-"` // by Key, set by validShelves: no word of a shelf is the realm's
}

// shelfDef is what gnoweb says of a shelf, and the floors it shows it from.
// The realm's own floors are permanent for its path; these are a gnoweb
// release away from rising with the chain, like the thresholds of earned
// reputation (ADR-004 §2, §5).
type shelfDef struct {
	Title  string
	Note   string
	Empty  string // what a pinned shelf says when it holds no app
	Pinned bool   // always drawn, from its own head
	// ownHead draws a shelf from its own head like a pinned one, though
	// only when it is long enough: a ranking shows its true top, even of
	// apps shown above.
	ownHead bool
	// minRanked and minWeek are the ranked stars, in all and over the last
	// week, an app needs to be on the shelf.
	minRanked, minWeek int
}

// Launch floors, the realm's own: raising one needs no transaction.
const (
	topStars      = 3
	trendingStars = 2
)

// homeShelfDefs and buildShelfDefs are the shelves this gnoweb draws, by the
// realm's key. A shelf it does not know is left out.
var (
	homeShelfDefs = map[string]shelfDef{
		"new": {Title: "New", Pinned: true, Empty: "No new app yet. List yours from a registered namespace: it opens here."},
		"top": {
			Title: "Top", Note: "Ranked by stars from accounts with a registered name.", Pinned: true, minRanked: topStars,
			Empty: "No app has " + strconv.Itoa(topStars) + " stars from named accounts yet. Found one you like? Star it: that's how apps get here.",
		},
		"trending":   {Title: "Trending this week", minWeek: trendingStars},
		"updated":    {Title: "Recently updated"},
		"rediscover": {Title: "Rediscover", Note: "Older apps, a new set every day."},
		"core":       {Title: "gno.land essentials", Note: "Built by the core team. Good first stops.", Pinned: true, Empty: "No core app is listed."},
	}
	buildShelfDefs = map[string]shelfDef{
		"new":     {Title: "New to build with"},
		"updated": {Title: "Recently updated"},
		"starred": {Title: "Most starred"},
	}
)

// listRef names a list of the store front ("top", "latest", "trending",
// "updated") and counts it; its title is gnoweb's own (lists).
type listRef struct {
	Key   string `json:"key"`
	Count int    `json:"count"`
	Title string `json:"-"`
}

// pulse counts store activity over the last 7 and 30 days. Other counts
// the realm may send are ignored.
type pulse struct {
	Listed7d  int `json:"listed_7d"`
	Stars7d   int `json:"stars_7d"`
	Listed30d int `json:"listed_30d"`
	Stars30d  int `json:"stars_30d"`
}

// builder aggregates a namespace's listings.
type builder struct {
	Namespace string `json:"namespace"`
	Listings  int    `json:"listings"`
	Stars     int    `json:"stars"`
}

// chainClock is the realm's view of the chain when it answered.
type chainClock struct {
	Height int64 `json:"height"`
	Time   int64 `json:"time"` // unix seconds
}

type homeResponse struct {
	envelope
	chainClock
	Pulse      pulse      `json:"pulse"`
	Categories []category `json:"categories"`
	Shelves    []shelf    `json:"shelves"`
	Pick       string     `json:"pick"` // the app GovDAO picked, if any
	Listings   []listing  `json:"listings"`
	// MovedTo is the store GovDAO named as this one's successor, a realm
	// path on this chain, if any; the web path once validated.
	MovedTo string `json:"moved_to"`
}

type buildResponse struct {
	envelope
	MovedTo  string  `json:"moved_to"` // as homeResponse's
	Shelves  []shelf `json:"shelves"`
	Builders struct {
		Top []builder `json:"top"`
		New []builder `json:"new"`
	} `json:"builders"`
	Listings []listing `json:"listings"`
}

// pageResponse is one page of a category or of a list: the realm echoes
// which one under its own field name.
type pageResponse struct {
	envelope
	Category category  `json:"category"`
	List     listRef   `json:"list"`
	Page     int       `json:"page"`
	Pages    int       `json:"pages"`
	Listings []listing `json:"listings"`
}

// decode fetches one api/v1 endpoint and decodes it into dst. A failure is
// logged here, once per load (the cache keeps failures briefly), not once
// per reader; a missing entry is not a failure.
func (h *Handler) decode(ctx context.Context, endpoint string, dst interface{ check() error }) (err error) {
	defer func() { h.logFailure(endpoint, err) }()
	body, err := h.deps.Client.Realm(ctx, h.deps.RealmPath, "api/v1/"+endpoint)
	if err != nil {
		return err
	}
	if len(body) > maxPayload {
		return errTooLarge
	}
	if err = json.Unmarshal(body, dst); err != nil {
		return fmt.Errorf("store: decode %s: %w", endpoint, err)
	}
	return dst.check()
}

// logFailure logs a failed realm query; a missing entry is not a failure.
func (h *Handler) logFailure(endpoint string, err error) {
	if err != nil && !errors.Is(err, errNotFound) {
		h.deps.Logger.Warn("store: realm query failed", "endpoint", endpoint, "error", err)
	}
}

// The endpoint accessors below cache the decoded and sanitised response, so
// a cache hit costs neither an RPC nor a decode.

func (h *Handler) home(ctx context.Context) (*homeResponse, error) {
	return h.cache.home.Get(ctx, "home", func(ctx context.Context) (*homeResponse, error) {
		var res homeResponse
		if err := h.decode(ctx, "home", &res); err != nil {
			return nil, err
		}
		res.Time = max(res.Time, 0) // drives the Spotlight's window
		res.MovedTo = h.realmWebPath(res.MovedTo)
		res.Categories = slices.DeleteFunc(res.Categories, func(c category) bool { return !validCategory(c) })
		res.Listings = h.sanitize(res.Listings)
		res.Shelves = validShelves(res.Shelves, homeShelfDefs)
		applyFloors(res.Shelves, indexListings(res.Listings))
		return &res, nil
	})
}

func (h *Handler) build(ctx context.Context) (*buildResponse, error) {
	return h.cache.build.Get(ctx, "build", func(ctx context.Context) (*buildResponse, error) {
		var res buildResponse
		if err := h.decode(ctx, "build", &res); err != nil {
			return nil, err
		}
		res.MovedTo = h.realmWebPath(res.MovedTo)
		res.Shelves = validShelves(res.Shelves, buildShelfDefs)
		res.Listings = h.sanitize(res.Listings)
		res.Builders.Top = h.validBuilders(res.Builders.Top)
		res.Builders.New = h.validBuilders(res.Builders.New)
		return &res, nil
	})
}

// paged loads one page of a category or of a list, refusing an answer for
// another key or page.
func (h *Handler) paged(ctx context.Context, kind, key string, page int) (*pageResponse, error) {
	endpoint := kind + "/" + key + "/" + strconv.Itoa(page)
	return h.cache.pages.Get(ctx, endpoint, func(ctx context.Context) (*pageResponse, error) {
		var res pageResponse
		if err := h.decode(ctx, endpoint, &res); err != nil {
			return nil, err
		}
		echoed := res.List.Key
		if kind == pagedCategory {
			echoed = res.Category.Key
		}
		if echoed != key || res.Page != page || res.Pages < page {
			err := fmt.Errorf("store: %s answered %s/%d of %d", endpoint, echoed, res.Page, res.Pages)
			h.logFailure(endpoint, err)
			return nil, err
		}
		res.Listings = h.sanitize(res.Listings)
		if kind == pagedList && key == trendingKey {
			// The list behind Trending, held to the shelf's floor.
			res.Listings = slices.DeleteFunc(res.Listings, func(l listing) bool { return l.Week < trendingStars })
		}
		return &res, nil
	})
}

// The validators below keep what the templates may show from the realm's
// non-listing fields, dropping anything malformed: the realm is untrusted.

// validCategory keeps a category with a key the store's paths can carry, a
// displayable name and counts that hold together.
func validCategory(c category) bool {
	return reKey.MatchString(c.Key) && c.Label != "" && displayable(c.Label, maxTitleRunes) &&
		c.Confirmed >= 0 && c.Confirmed <= c.Count
}

// validShelves keeps the shelves defs knows, with what gnoweb says of each,
// and drops from them any bad list reference, which then just does not show.
func validShelves(in []shelf, defs map[string]shelfDef) []shelf {
	out := in[:0]
	for _, s := range in {
		def, ok := defs[s.Key]
		if !ok {
			continue
		}
		s.shelfDef = def
		refs := s.More[:0]
		for _, l := range s.More {
			if name, ok := namedList(l.Key); ok && l.Count >= 0 {
				l.Title = name.title
				refs = append(refs, l)
			}
		}
		s.More = refs
		out = append(out, s)
	}
	return out
}

// applyFloors drops from each shelf the apps below gnoweb's floors for it
// (shelfDef), by each listing's own counts.
func applyFloors(shelves []shelf, bySlug map[string]*listing) {
	for i := range shelves {
		s := &shelves[i]
		if s.minRanked == 0 && s.minWeek == 0 {
			continue
		}
		s.Slugs = slices.DeleteFunc(s.Slugs, func(slug string) bool {
			l := bySlug[slug]
			return l == nil || l.Ranked < s.minRanked || l.Week < s.minWeek
		})
	}
}

// validBuilders drops malformed rows, and the namespaces the operator
// denies.
func (h *Handler) validBuilders(in []builder) []builder {
	return slices.DeleteFunc(in, func(b builder) bool {
		return !reNamespace.MatchString(b.Namespace) || b.Listings < 0 || b.Stars < 0 || h.denied(b.Namespace)
	})
}
