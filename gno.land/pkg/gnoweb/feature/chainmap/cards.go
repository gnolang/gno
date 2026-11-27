package chainmap

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

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
	// Name is the package's name, from its package clause.
	Name string
	// Doc is the package's doc comment, as markdown.
	Doc string
	// Actions are its exported crossing functions: what a reader can call.
	Actions []string
	// Height is the block it was deployed in; 0 is genesis, and
	// UnknownHeight when the node could not say.
	Height int
}

// UnknownHeight is a deploy height the node could not give, such as on a node
// without vm/qpkgmeta_json.
const UnknownHeight = -1

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

	// activityMaxAge is how old the week's tip may be for a card to show its
	// activity. The aggregate is refreshed every activityTTL, so a tip older
	// than this means the indexer, or the chain, stopped some time ago, and
	// "the last 7 days" no longer ends now.
	activityMaxAge = 3 * activityTTL
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

// newSummaryFlight keeps genesis packages over the others when full: the
// chain bounds those, while anyone can deploy enough packages to push out
// the rest.
func newSummaryFlight() *flight[cardSummary] {
	return &flight[cardSummary]{
		ttl: summaryTTL, errTTL: failureTTL, timeout: summaryTimeout, max: cardCacheMax, stale: true,
		keep: func(s cardSummary) bool { return s.height == 0 },
	}
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

	// The week's activity is read beside the packages and never waited on
	// past them: once they are in, the cards show the aggregate the flight
	// holds, and a scan this starts carries on for the next reader. Only
	// realms are called, so a listing of pure packages never asks.
	activity := h.activity != nil && realms
	if activity {
		// Detached from the listing, so the scan starts even when every
		// card is warm and the listing returns at once.
		actx, acancel := context.WithTimeout(context.WithoutCancel(ctx), cardsWait)
		go func() {
			defer acancel()
			h.startActivity(actx)
		}()
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

	if !activity {
		return cards
	}
	if week, ok := h.activity.held(activityKey); ok && time.Since(week.Until) < activityMaxAge {
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
		return cardSummary{description: describe(p.Doc, p.Name), actions: len(p.Actions), height: p.Height}, nil
	})
	switch {
	case err == nil:
	case errors.Is(err, ErrPending), errors.Is(err, ErrNotLive), errors.Is(err, errTransient):
		return
	case errors.Is(err, errFetchPanicked):
		h.deps.Logger.Error("cards: package summary panicked", "path", c.Path, "error", err)
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

// startActivity starts the week's scan unless a fresh aggregate is held, or
// joins the one running, and logs a failure. Cards read the aggregate from
// the flight, so they never wait on the scan.
func (h *Handler) startActivity(ctx context.Context) {
	if _, err := h.weekActivity(ctx); err != nil && !errors.Is(err, ErrPending) {
		h.deps.Logger.Warn("cards: activity unavailable", "error", err)
	}
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

// describe is a card's description: the opening of the doc comment, plain
// text, in the author's words and case, less the Go doc clause naming the
// package, "Package foo is a", which the card's title already says. A doc
// comment that opens with something else, such as a disclaimer, and names
// the package in a later paragraph is read from that paragraph.
func describe(doc, name string) string {
	if name == "" {
		return markdown.Summary(doc, descriptionMaxRunes)
	}
	clause := regexp.MustCompile(`^(?i)package\s+` + regexp.QuoteMeta(name) + `\s+(?:(?:is|are)\s+)?(?:(?:an?|the)\s+)?`)
	paras := strings.Split(doc, "\n\n")
	for i, p := range paras {
		if clause.MatchString(p) {
			doc = strings.Join(paras[i:], "\n\n")
			break
		}
	}
	return clause.ReplaceAllString(markdown.Summary(doc, descriptionMaxRunes), "")
}
