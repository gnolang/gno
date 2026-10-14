// Package store is the on-chain registry behind gnoweb's Explore
// (gno.land/adr/adr-004-gnoweb-store.md).
//
// A listing is identified by a store-wide slug. An app lists itself:
// Register records the calling realm's own path, so deploying the realm is
// the proof of control. Claim lets a namespace owner
// list packages, services and realms that cannot call Register. Render
// serves a markdown store and a versioned JSON API (api/v1/...) for gnoweb.
package store

import (
	"chain"
	"strconv"
	"time"

	"gno.land/p/nt/avl/v0"
	"gno.land/r/sys/names"
	"gno.land/r/sys/users"
)

// Durations are measured in block time, so they mean what they say whatever
// the chain's block rate; heights only order. blockTime is the one clock.
const (
	day = 24 * 60 * 60 // seconds

	newWindow   = 14 * day // an app is new for Rediscover and "New to build with"
	epochLength = day      // rotation and counter period
	richDelay   = day      // publication delay of rich fields (T1, T8)
	// Launch values for a young chain (ADR §2): earned reputation only gates
	// owner assets and the Spotlight, so it may be reachable early. They are
	// constants: a realm upgrade raises them as the chain grows.
	earnAge      = 14 * day // minimum listing age for earned reputation
	starMaturity = 3 * day  // a star counts towards reputation after this
	earnStars    = 5        // qualifying stars needed for earned reputation
	// nameMaturity is how long the store must have known an address's name
	// before its stars rank: names are free on a young chain, so a name made
	// for one campaign has to wait (T8).
	nameMaturity = 3 * day
	// confirmStars ranked stars confirm a claimed listing (T11).
	confirmStars = 3
	// updateCooldown is the least time between two "updated" announcements of
	// a listing, so that re-registering in a loop buys no visibility.
	updateCooldown = 7 * epochLength

	// No global cap: reads are bounded by shelf sizes, trendingScan and
	// pagination, and storage deposits price every write.
	maxPerNamespace  = 5  // visible apps, services and packages together
	maxStarsPerEpoch = 20 // per address, against star bombing
)

func blockTime() int64 {
	return time.Now().Unix()
}

// voters holds what the store keeps per starring address.
var voters avl.Tree // address -> voter, a value: a pointer would store one more object per address

type voter struct {
	Epoch    int64 // of the last star
	Spent    int   // stars spent in Epoch
	NameSeen int64 // when the store first saw the address with a registered name, 0 if never
}

// Register lists the calling realm as an app. It must be cross-called by the
// app realm itself, typically from its init():
//
//	store.Register(cross(cur), "acme-swap", "Acme Swap", "Swap tokens in one click", "defi", 3)
//
// Calling Register again, as an upgraded realm's init does, updates the
// listing: the original slug and creation time are kept and the slug
// argument is ignored. A real change is announced as an update at most once
// a week. Register never panics, but the deploying transaction pays for what
// it stores here (about 2 GNOT for a namespace's first listing), so the
// deploy's -max-deposit must cover it. A refused call emits
// StoreRegisterRejected with the path and a short reason.
// The namespace owner can then list the realm with Claim, and the realm's
// next Register confirms it.
func Register(cur realm, slug, title, tagline, category string, palette int) {
	// The weekly announcement limit is updateCooldown.
	path := cur.Previous().PkgPath()
	kind := kindApp
	if existing := lookup(&byPath, path); existing != nil {
		kind = existing.Kind // a service claimed by its owner stays one
	}
	l := newListing(path, slug, title, tagline, category, palette, kind)
	l.Confirmed = true // deploying the realm proves it exists
	if reason := upsert(l, blockTime()); reason != "" {
		chain.Emit("StoreRegisterRejected", "path", path, "reason", reason)
		return
	}
	chain.Emit("StoreRegistered", "path", path, "slug", lookup(&byPath, path).Slug)
}

// Claim lists, or updates, path on behalf of its namespace owner: a /p/
// package (kind "package"), or a realm that cannot call Register (kind "app"
// or "service"). You must own the path's namespace. The store cannot check
// that a package exists, so a claimed listing is unconfirmed until it has 3
// stars that count toward its rank, or the realm's own Register confirms it.
// Until then a claimed package or service shows nowhere, and a claimed app
// only on its category page; once confirmed, it reaches the shelves, the
// activity feed and the builders. Unlike Register, Claim must be called
// directly, and panics with the reason when it refuses.
func Claim(cur realm, path, slug, title, tagline, category string, palette int, kind string) {
	// Ownership is ownsNamespace; a claimed path is not verified beyond its
	// grammar, so it takes confirmStars ranked stars to confirm (T11).
	caller := cur.Previous()
	if !caller.IsUserCall() {
		panic("store: claims must come from a direct user call")
	}
	ns := namespaceOf(path)
	if ns == "" || !ownsNamespace(caller.Address(), ns) {
		panic("store: caller does not own the namespace")
	}
	if l := lookup(&byPath, path); l != nil && l.core {
		panic("store: a core listing cannot be claimed") // frozen in code (ADR §4)
	}
	if reason := upsert(newListing(path, slug, title, tagline, category, palette, kind), blockTime()); reason != "" {
		panic("store: " + reason)
	}
	chain.Emit("StoreClaimed", "path", path, "slug", lookup(&byPath, path).Slug, "kind", kind)
}

// ownsNamespace reports whether addr owns namespace ns: its own address, or
// a name it currently holds in r/sys/users, as r/sys/names decides deploys.
// The store asks r/sys/users itself because r/sys/names answers no for
// everyone while GovDAO pauses deploys, which would make owners strangers to
// Star and refuse every Claim. Until a chain enables r/sys/names anyone may
// deploy under a name (dev chains, tests), so only an address counts then:
// one user can never edit another's listing.
func ownsNamespace(addr address, ns string) bool {
	if addr.String() == ns {
		return true
	}
	if !names.IsEnabled() {
		return false
	}
	u, current := users.ResolveName(ns)
	return u != nil && current && u.Addr() == addr
}

func newListing(path, slug, title, tagline, category string, palette int, kind string) *listing {
	return &listing{
		Path:     path,
		Slug:     slug,
		Kind:     kind,
		Title:    title,
		Tagline:  tagline,
		Category: category,
		Palette:  palette,
	}
}

// SubmitRich proposes a new icon and cover for the calling realm's listing:
// only the listed realm itself may call it, from its own code. The change
// shows a day later; until then the store keeps showing the previous ones.
func SubmitRich(cur realm, icon, cover string) {
	// The delay is richDelay (T1, T8). Owner functions are authorised by
	// realm path, not by an address: Register usually runs from the app's
	// init, where no user identity worth keeping as owner is available.
	if paused {
		panic("store: paused")
	}
	l := byPathOrPanic(cur.Previous().PkgPath())
	r := rich{Icon: icon, Cover: cover}
	if reason := validateRich(r); reason != "" {
		panic("store: " + reason)
	}
	l.submitRich(r, blockTime())
	chain.Emit("StoreRichSubmitted", "slug", l.Slug, "path", l.Path)
}

// Star an app you like. One star per app, up to 20 new stars a day. Your
// star counts toward an app's rank once your address has a registered @name
// the store has seen for 3 days; until then it shows in the app's total, and
// starring the app again once it would count makes it count. Stars on apps
// in your own namespace never rank.
func Star(cur realm, slug string) {
	// Only direct user calls are accepted (starTarget), so a `maketx run`
	// script or another realm cannot star on a user's behalf. There are no
	// downvotes: they would be a bombing weapon. maxStarsPerEpoch bounds
	// star bombing per address.
	//
	// The star ranks the listing only if the caller has a registered name the
	// store has known for nameMaturity, and does not own the listing's
	// namespace: a free address, a name made for one campaign or a vote for
	// oneself only adds to the displayed count (T8). r/sys/users keeps no
	// registration height, so a name's first star only starts the clock.
	l, caller := starTarget(cur, slug)
	addr := caller.String()
	now := blockTime()
	named := users.ResolveAddress(caller) != nil
	nameSeen := spendStar(addr, now, named)
	ranked := named && now-nameSeen >= nameMaturity && !ownsNamespace(caller, l.Namespace)
	if l.starred(addr) {
		// Starring again is how a star given before the name counted comes
		// to rank: the first one only starts the name's clock.
		if l.starOf(addr).Ranked || !ranked {
			panic("store: already starred")
		}
		l.removeStar(addr, now)
	}
	l.addStar(addr, star{At: now, Ranked: ranked})
	chain.Emit("StoreStarred", "slug", slug, "ranked", strconv.FormatBool(ranked))
}

// Unstar removes your star from an app. It does not give back that day's
// star: the 20 a day count the stars you add.
func Unstar(cur realm, slug string) {
	l, caller := starTarget(cur, slug)
	if !l.starred(caller.String()) {
		panic("store: not starred")
	}
	l.removeStar(caller.String(), blockTime())
	chain.Emit("StoreUnstarred", "slug", slug)
}

func starTarget(cur realm, slug string) (*listing, address) {
	caller := cur.Previous()
	if !caller.IsUserCall() {
		panic("store: stars must come from a direct user call")
	}
	return bySlugOrPanic(slug), caller.Address()
}

// spendStar charges one star to addr's budget for the epoch of now, records
// now as the time the store first saw addr named if it is the first, and
// returns that time (0 if never).
func spendStar(addr string, now int64, named bool) int64 {
	v, _ := voters.Get(addr).(voter)
	if epoch := now / epochLength; v.Epoch != epoch {
		v.Epoch, v.Spent = epoch, 0
	}
	if v.Spent >= maxStarsPerEpoch {
		panic("store: star rate limit")
	}
	v.Spent++
	if named && v.NameSeen == 0 {
		v.NameSeen = now
	}
	voters.Set(addr, v)
	return v.NameSeen
}
