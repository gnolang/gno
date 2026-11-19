package chainmap

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"golang.org/x/sync/errgroup"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/markdown"
)

// PackageReader reads what a package says about itself from the chain.
type PackageReader interface {
	// Summary reads the live package at pkgPath, a gnoweb path such as
	// "/r/gnoland/blog". It returns ErrNotLive when none sits there.
	Summary(ctx context.Context, pkgPath string) (PackageSummary, error)
}

// PackageSummary is what the chain says of a package, for its card.
type PackageSummary struct {
	// Doc is the package's doc comment, as markdown.
	Doc string
	// Actions are its exported crossing functions: what a reader can call.
	Actions []string
	// Height is the block it was deployed in; 0 is genesis.
	Height int
}

// Card waits and budgets.
const (
	// cardsWait is how long a listing waits for its cards. A card not ready
	// by then shows its path and links, and its fetch carries on for the
	// next reader, so a cold listing fills in over a few loads instead of
	// holding the page.
	cardsWait = 1500 * time.Millisecond

	// cardConcurrency bounds the goroutines one listing runs at once.
	cardConcurrency = 8

	// maxCardFetches bounds the card reads running at once across all
	// readers, chain and indexer together: a quarter of the RPC client's
	// default cap, half of the indexer client's. A fetch outlives the listing
	// that started it, and many cold listings at once must not take the
	// slots every other page needs. A card refused a slot stays blank, and a
	// later reader retries.
	maxCardFetches = 8

	// summaryTTL is how long a package's card is served fresh; past it, the
	// card is served while it is read again. It is short because a package
	// can change under the same path: a private realm redeployed, or any
	// realm reloaded by gnodev.
	summaryTTL = time.Minute

	// summaryTimeout bounds one package's reads, whoever started them.
	summaryTimeout = 10 * time.Second

	// cardCacheMax bounds the packages and blocks remembered.
	cardCacheMax = 4096

	// blockTimeTTL is how long a block's time is kept: it never changes, so
	// the TTL only lets the cache forget heights nobody asks for.
	blockTimeTTL = 24 * time.Hour

	// descriptionMaxRunes caps a card's description at about two lines.
	descriptionMaxRunes = 240
)

// errCardsBusy is returned when every card fetch slot is taken. Transient, so
// never remembered.
var errCardsBusy = fmt.Errorf("%w: every card fetch slot is taken", errTransient)

// cardSummary is a package's card as cached: the doc comment already reduced
// to its description, so a listing never parses markdown.
type cardSummary struct {
	description string
	actions     int
	height      int
}

func newSummaryFlight() *flight[cardSummary] {
	return &flight[cardSummary]{ttl: summaryTTL, errTTL: failureTTL, timeout: summaryTimeout, max: cardCacheMax, stale: true}
}

func newBlockTimeFlight() *flight[time.Time] {
	return &flight[time.Time]{ttl: blockTimeTTL, errTTL: failureTTL, timeout: summaryTimeout, max: cardCacheMax}
}

// Cards describes each package of a listing, in the order given. Whatever is
// not ready within cardsWait is left blank, never guessed: a card always has
// its path and links.
func (h *Handler) Cards(ctx context.Context, paths []string) []components.PackageCard {
	cards := make([]components.PackageCard, len(paths))
	realms := false
	for i, p := range paths {
		cards[i] = components.PackageCard{Path: p, Realm: strings.HasPrefix(p, "/r/")}
		realms = realms || cards[i].Realm
	}
	if h.summaries == nil {
		return cards
	}

	wctx, cancel := context.WithTimeout(ctx, cardsWait)
	defer cancel()

	// The week's activity is read beside the packages, not before them: a
	// cold scan must not spend the listing's whole wait. Only realms are
	// called, so a listing of pure packages never asks.
	weekc := make(chan *Activity, 1)
	if h.activity != nil && realms {
		go func() { weekc <- h.cardActivity(wctx) }()
	} else {
		weekc <- nil
	}

	var g errgroup.Group
	g.SetLimit(cardConcurrency)
	for i := range cards {
		c := &cards[i]
		g.Go(func() error {
			h.fillCard(wctx, c)
			return nil
		})
	}
	_ = g.Wait() // no task returns an error

	if week := <-weekc; week != nil {
		for i := range cards {
			if cards[i].Realm {
				fillActivity(&cards[i], week, h.deps.Domain+cards[i].Path)
			}
		}
	}
	return cards
}

// fillCard sets what the chain says about c's package, and when an indexer
// can date it, its deploy date.
func (h *Handler) fillCard(ctx context.Context, c *components.PackageCard) {
	s, err := h.summaries.get(ctx, c.Path, func(ctx context.Context) (cardSummary, error) {
		release, err := h.takeCardSlot()
		if err != nil {
			return cardSummary{}, err
		}
		defer release()
		p, err := h.deps.Packages.Summary(ctx, c.Path)
		if err != nil {
			return cardSummary{}, err
		}
		return cardSummary{description: describe(p.Doc), actions: len(p.Actions), height: p.Height}, nil
	})
	switch {
	case err == nil:
	case errors.Is(err, ErrPending), errors.Is(err, ErrNotLive), errors.Is(err, errTransient):
		return
	default:
		h.deps.Logger.Debug("cards: package unreadable", "path", c.Path, "error", err)
		return
	}

	c.Known, c.Block, c.Description = true, s.height, s.description
	if c.Realm {
		c.Actions = s.actions
	}
	if c.Block > 0 && h.blockTimes != nil {
		c.Date = h.blockTime(ctx, c.Block)
	}
}

// takeCardSlot admits one card fetch, chain or indexer, or refuses at once
// when maxCardFetches already run: cards never queue on the backends other
// pages need. The returned func gives the slot back.
func (h *Handler) takeCardSlot() (release func(), err error) {
	select {
	case h.cardSlots <- struct{}{}:
		return func() { <-h.cardSlots }, nil
	default:
		return nil, errCardsBusy
	}
}

// cardActivity is the week's aggregate, or nil when it is not ready in time
// or the indexer failed: cards then show no activity.
func (h *Handler) cardActivity(ctx context.Context) *Activity {
	a, err := h.weekActivity(ctx)
	if err != nil {
		if !errors.Is(err, ErrPending) {
			h.deps.Logger.Warn("cards: activity unavailable", "error", err)
		}
		return nil
	}
	return a
}

// blockTime is when height was produced, or the zero time when the indexer
// cannot say in time.
func (h *Handler) blockTime(ctx context.Context, height int) time.Time {
	t, err := h.blockTimes.get(ctx, strconv.Itoa(height), func(ctx context.Context) (time.Time, error) {
		release, err := h.takeCardSlot()
		if err != nil {
			return time.Time{}, err
		}
		defer release()
		b, err := h.deps.Indexer.Block(ctx, height)
		if err != nil {
			return time.Time{}, err
		}
		return b.Time, nil
	})
	if err != nil {
		return time.Time{}
	}
	return t
}

// fillActivity sets the realm's use over the window. A realm nobody called
// shows nothing: under a partial window that zero would say nothing.
func fillActivity(c *components.PackageCard, a *Activity, pkgPath string) {
	users := a.Callers[pkgPath]
	if users == 0 {
		return
	}
	c.WeeklyUsers, c.AtLeast = users, a.Partial
	if last, ok := a.Last[pkgPath]; ok {
		c.LastUsed = a.timeOf(last)
	}
}

// packageClause matches the Go doc convention a doc comment opens with,
// "Package foo is a", which a card under the path already says.
var packageClause = regexp.MustCompile(`^(?i:package)\s+\S+\s+(?:(?:is|are)\s+)?(?:(?:an?|the)\s+)?`)

// describe is a card's description: the opening of the doc comment, plain
// text, without the package clause.
func describe(doc string) string {
	s := markdown.Summary(doc, descriptionMaxRunes)
	s = packageClause.ReplaceAllString(s, "")
	r, size := utf8.DecodeRuneInString(s)
	if size == 0 {
		return ""
	}
	return string(unicode.ToUpper(r)) + s[size:]
}
