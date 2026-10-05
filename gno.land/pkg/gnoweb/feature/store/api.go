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

type category struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Count int    `json:"count"`
}

// shelf is a section of the front page or the build lens, in the realm's
// order.
type shelf struct {
	Title   string    `json:"title"`
	Note    string    `json:"note"`
	Pinned  bool      `json:"pinned"` // always drawn, from its own head
	Empty   string    `json:"empty"`  // what a pinned shelf says when it holds no app
	Slugs   []string  `json:"slugs"`
	Stars7d []int     `json:"stars_7d"` // per slug, its ranked stars this week, if given
	More    []listRef `json:"more"`     // the paged lists behind the shelf
}

// listRef names a list of the store front ("top", "latest", "trending",
// "updated").
type listRef struct {
	Key   string `json:"key"`
	Title string `json:"title"`
	Count int    `json:"count"`
}

// pulse counts store activity over the last 7 and 30 days. Other counts
// the realm may send are ignored.
type pulse struct {
	Listed7d  int `json:"listed_7d"`
	Stars7d   int `json:"stars_7d"`
	Listed30d int `json:"listed_30d"`
	Stars30d  int `json:"stars_30d"`
}

// activity is one store event: "listed", "updated" or "milestone".
type activity struct {
	Kind  string `json:"kind"`
	Slug  string `json:"slug"`
	Stars int    `json:"stars"`
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
	Activity   []activity `json:"activity"`
	Pick       string     `json:"pick"` // the app GovDAO picked, if any
	Listings   []listing  `json:"listings"`
}

type buildResponse struct {
	envelope
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
	defer func() {
		if err != nil && !errors.Is(err, errNotFound) {
			h.deps.Logger.Warn("store: realm query failed", "endpoint", endpoint, "error", err)
		}
	}()
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

// The endpoint accessors below cache the decoded and sanitised response, so
// a cache hit costs neither an RPC nor a decode.

func (h *Handler) home(ctx context.Context) (*homeResponse, error) {
	return cached(ctx, h.cache, "home", func(ctx context.Context) (*homeResponse, error) {
		var res homeResponse
		if err := h.decode(ctx, "home", &res); err != nil {
			return nil, err
		}
		res.Height = max(res.Height, 0) // drives the Spotlight's window
		res.Categories = slices.DeleteFunc(res.Categories, func(c category) bool { return !validRef(c.Key, c.Label, c.Count) })
		res.Shelves = validShelves(res.Shelves)
		res.Listings = h.sanitize(res.Listings)
		return &res, nil
	})
}

func (h *Handler) build(ctx context.Context) (*buildResponse, error) {
	return cached(ctx, h.cache, "build", func(ctx context.Context) (*buildResponse, error) {
		var res buildResponse
		if err := h.decode(ctx, "build", &res); err != nil {
			return nil, err
		}
		res.Shelves = validShelves(res.Shelves)
		res.Listings = h.sanitize(res.Listings)
		res.Builders.Top = validBuilders(res.Builders.Top)
		res.Builders.New = validBuilders(res.Builders.New)
		return &res, nil
	})
}

// paged loads one page of a category or of a list, refusing an answer for
// another key or page.
func (h *Handler) paged(ctx context.Context, kind, key string, page int) (*pageResponse, error) {
	endpoint := kind + "/" + key + "/" + strconv.Itoa(page)
	return cached(ctx, h.cache, endpoint, func(ctx context.Context) (*pageResponse, error) {
		var res pageResponse
		if err := h.decode(ctx, endpoint, &res); err != nil {
			return nil, err
		}
		echoed := res.List.Key
		if kind == pagedCategory {
			echoed = res.Category.Key
		}
		if echoed != key || res.Page != page || res.Pages < page {
			return nil, fmt.Errorf("store: %s answered %s/%d of %d", endpoint, echoed, res.Page, res.Pages)
		}
		res.Listings = h.sanitize(res.Listings)
		return &res, nil
	})
}

// The validators below keep what the templates may show from the realm's
// non-listing fields, dropping anything malformed: the realm is untrusted.

// validRef keeps a category or a list reference: a key the store's paths
// can carry, a displayable name, and a count.
func validRef(key, label string, count int) bool {
	return reKey.MatchString(key) && label != "" && displayable(label, maxTitleRunes) && count >= 0
}

// validShelves drops shelves with a bad title, and from the others any bad
// note, weekly counts or list reference, which then just do not show.
func validShelves(in []shelf) []shelf {
	out := in[:0]
	for _, s := range in {
		if s.Title == "" || !displayable(s.Title, maxTitleRunes) {
			continue
		}
		if !displayable(s.Note, maxTaglineRunes) {
			s.Note = ""
		}
		if !displayable(s.Empty, maxEmptyRunes) {
			s.Empty = ""
		}
		if len(s.Stars7d) != len(s.Slugs) || slices.ContainsFunc(s.Stars7d, func(n int) bool { return n < 0 }) {
			s.Stars7d = nil
		}
		s.More = slices.DeleteFunc(s.More, func(l listRef) bool { return !validRef(l.Key, l.Title, l.Count) })
		out = append(out, s)
	}
	return out
}

func validBuilders(in []builder) []builder {
	return slices.DeleteFunc(in, func(b builder) bool {
		return !reNamespace.MatchString(b.Namespace) || b.Listings < 0 || b.Stars < 0
	})
}
