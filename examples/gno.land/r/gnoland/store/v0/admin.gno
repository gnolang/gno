package store

import (
	"chain"
	"chain/runtime"
	"errors"
	"strconv"

	"gno.land/r/gov/dao"
)

// GovDAO governs the store (ADR §3, T10) through the proposals built here:
// it can pause new listings, hide listings and pick an app as the front
// page's hero, never edit them. A callback
// returns an error rather than panicking, so a proposal whose execution
// fails stays readable instead of stuck.

// paused refuses new listings and listing updates.
var paused bool

// pick is the app GovDAO features as the front page's hero, until a height;
// the only editorial act in the store, voted, labelled and bounded in time.
// It is bound to the title the app had when picked: a rename ends it
// (dropPick), and picked checks the title too, so a renamed app is never
// shown as the one voted on.
var pick struct {
	listing *listing
	title   string // titleKey of the app's title when picked
	until   int64
}

// maxPickDays bounds a pick, so none outlives the vote that made it.
const maxPickDays = 30

// hideReasons is the closed set of reasons for hiding a listing.
var hideReasons = []string{"impersonation", "phishing", "broken", "offensive", "other"}

var (
	errUnknownListing = errors.New("store: unknown listing")
	errNotAnApp       = errors.New("store: only an app can be picked")
	errNotConfirmed   = errors.New("store: only a confirmed app can be picked")
	errSlugMoved      = errors.New("store: the slug now names another listing")
	errTitleChanged   = errors.New("store: the app changed its title since the proposal")
	errBadPickDays    = errors.New("store: a pick lasts 1 to " + strconv.Itoa(maxPickDays) + " days")
	errUnknownReason  = errors.New("store: unknown hide reason")
	errNotHidden      = errors.New("store: not hidden")
)

// ProposePause returns a GovDAO proposal that pauses new listings and
// listing updates: Register is refused (it emits StoreRegisterRejected) and
// Claim panics. Stars, SubmitRich and hiding keep working.
func ProposePause(cur realm) dao.ProposalRequest {
	return proposal(cur, "Pause the store", "Pauses new listings and listing updates in the store.",
		func() error { paused = true; return nil })
}

// ProposeUnpause returns a GovDAO proposal that lifts a pause.
func ProposeUnpause(cur realm) dao.ProposalRequest {
	return proposal(cur, "Unpause the store", "Lets the store take new listings and listing updates again.",
		func() error { paused = false; return nil })
}

// ProposeHide returns a GovDAO proposal that hides the listing slug for
// reason, one of hideReasons. Both are checked now, so a proposal names a
// listing that exists and a known reason. It is bound to that listing's
// path: if the slug names another listing by execution, it fails.
func ProposeHide(cur realm, slug, reason string) dao.ProposalRequest {
	if err := checkHide(slug, reason); err != nil {
		panic(err)
	}
	path := lookup(&listings, slug).Path
	return proposal(cur, "Hide "+slug+" from the store", "Hides the store listing "+slug+" ("+path+"), for "+reason+".",
		func() error { return hideBound(slug, path, reason) })
}

// ProposeUnhide returns a GovDAO proposal that restores the hidden listing
// of path. It takes the path because a hidden listing's slug may since
// belong to another one.
func ProposeUnhide(cur realm, path string) dao.ProposalRequest {
	l, err := hiddenListing(path)
	if err != nil {
		panic(err)
	}
	return proposal(cur, "Unhide "+l.Slug+" in the store", "Restores the hidden store listing of "+l.Path+".",
		func() error { return unhide(path) })
}

// ProposeSpotlight returns a GovDAO proposal that features the app slug as
// the store front page's hero for days days (1 to maxPickDays), labelled as
// picked by GovDAO. A later pick replaces it; hiding the app, or its ceasing
// to be a shown app, ends it. Like a hide, it is bound to the app's path,
// and also to its title: an app renamed before or after execution is not
// the one voted on.
func ProposeSpotlight(cur realm, slug string, days int) dao.ProposalRequest {
	if err := checkPick(slug, days); err != nil {
		panic(err)
	}
	l := lookup(&listings, slug)
	path, title := l.Path, titleKey(l.Title)
	n := strconv.Itoa(days)
	return proposal(cur, "Feature "+slug+" as the front page's hero for "+n+" days",
		"Features "+slug+" ("+path+") as the store front page's hero for "+n+" days, labelled as picked by GovDAO.",
		func() error { return pickBound(slug, path, title, days, runtime.ChainHeight()) })
}

// hideBound and pickBound execute a hide or a pick proposed for the listing
// of path under slug; a pick, for that listing under the title it had then
// (a titleKey).
func hideBound(slug, path, reason string) error {
	if err := checkSlugPath(slug, path); err != nil {
		return err
	}
	return hide(slug, reason)
}

func pickBound(slug, path, title string, days int, h int64) error {
	if err := checkSlugPath(slug, path); err != nil {
		return err
	}
	if titleKey(lookup(&listings, slug).Title) != title {
		return errTitleChanged
	}
	return setPick(slug, days, h)
}

// checkSlugPath checks that slug still names the visible listing of path. A
// hide releases the slug, so between a proposal and its execution it may
// pass to an unrelated listing, which must not be hidden or picked instead.
func checkSlugPath(slug, path string) error {
	l := lookup(&listings, slug)
	switch {
	case l == nil:
		return errUnknownListing
	case l.Path != path:
		return errSlugMoved
	}
	return nil
}

// checkPick checks that slug is an app picked() would show: visible and
// confirmed, so a bare claim is never picked.
func checkPick(slug string, days int) error {
	l := lookup(&listings, slug)
	switch {
	case l == nil:
		return errUnknownListing
	case l.Kind != kindApp:
		return errNotAnApp
	case !l.shownApp():
		return errNotConfirmed
	case days < 1 || days > maxPickDays:
		return errBadPickDays
	}
	return nil
}

// setPick picks slug from height h for days days, under its current title.
func setPick(slug string, days int, h int64) error {
	if err := checkPick(slug, days); err != nil {
		return err
	}
	l := lookup(&listings, slug)
	pick.listing, pick.title, pick.until = l, titleKey(l.Title), h+int64(days)*blocksPerDay
	chain.Emit("StorePicked", "slug", slug, "days", strconv.Itoa(days))
	return nil
}

// picked returns the app GovDAO picked, while the pick lasts and the app is
// still shown as an app under the title it was picked with, or nil.
func picked(h int64) *listing {
	if l := pick.listing; l != nil && h < pick.until && l.shownApp() && titleKey(l.Title) == pick.title {
		return l
	}
	return nil
}

// dropPick ends the pick of l, if it is the picked app. A hide or a change
// that takes l off the shown apps calls it, so that undoing the change does
// not bring back a pick GovDAO made for the app as it was.
func dropPick(l *listing) {
	if pick.listing == l {
		pick.listing = nil
	}
}

func proposal(cur realm, title, desc string, apply func() error) dao.ProposalRequest {
	return dao.NewProposalRequest(title, desc, dao.NewSimpleExecutor(0, cur, func(realm) error { return apply() }, ""))
}

// hide removes a listing from every shelf, category page and API response,
// and releases its slug, its title and its namespace slot so it blocks no
// one. The record stays under its path for audit, so unhide can restore it.
func hide(slug, reason string) error {
	if err := checkHide(slug, reason); err != nil {
		return err
	}
	l := lookup(&listings, slug)
	l.unindex()
	dropPick(l)
	listings.Remove(l.Slug)
	byTitle.Remove(titleKey(l.Title))
	l.hidden = true
	chain.Emit("StoreHidden", "slug", slug, "reason", reason)
	return nil
}

// unhide restores the hidden listing of path, unless its slug, its title or
// its namespace's slots were taken while it was hidden.
func unhide(path string) error {
	l, err := hiddenListing(path)
	if err != nil {
		return err
	}
	if reason := available(l); reason != "" {
		return errors.New("store: " + reason)
	}
	l.hidden = false
	listings.Set(l.Slug, l)
	byTitle.Set(titleKey(l.Title), l.Slug)
	l.index()
	chain.Emit("StoreUnhidden", "slug", l.Slug)
	return nil
}

// checkHide checks that slug is a visible listing and reason a known one.
func checkHide(slug, reason string) error {
	switch {
	case !isHideReason(reason):
		return errUnknownReason
	case lookup(&listings, slug) == nil: // visible listings only
		return errUnknownListing
	}
	return nil
}

// hiddenListing returns the hidden listing of path.
func hiddenListing(path string) (*listing, error) {
	l := lookup(&byPath, path)
	if l == nil || !l.hidden {
		return nil, errNotHidden
	}
	return l, nil
}

func isHideReason(reason string) bool {
	for _, r := range hideReasons {
		if r == reason {
			return true
		}
	}
	return false
}
