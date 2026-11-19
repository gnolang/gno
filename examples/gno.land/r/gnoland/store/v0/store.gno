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
	// earnStars is the ranked star whose time the API reports
	// (fifth_ranked). gnoweb decides earned reputation (ADR §2) from it, the
	// listing's age and its pending changes, so its thresholds can rise with
	// the chain: these constants are permanent for this path.
	earnStars = 5
	// nameMaturity is how long the store must have known an address's name
	// before its stars rank: names are free on a young chain, so a name made
	// for one campaign has to wait (T8).
	nameMaturity = 3 * day
	// confirmStars ranked stars confirm a claimed listing (T11).
	confirmStars = 3
	// updateCooldown is the least time between two changes of a listing
	// that count as updates, so that re-registering in a loop buys no
	// visibility.
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
// Calling Register again updates the listing: the original slug and
// creation time are kept and the slug argument is ignored, unless the store
// moved the listing's slug (StoreClaimMoved): it then takes the slug it asks
// for, once free. A real change counts as an update at most once a week. Register never panics, but the
// deploying transaction pays for what it stores here (about 2 GNOT for a
// namespace's first listing), so the deploy's -max-deposit must cover it. A
// refused call emits StoreRegisterRejected with the path, a short reason
// and, when another listing holds the slug or title, its slug (holder).
// An init runs once, so call Register from an exported function too, which
// anyone can call again once the cause is gone, or to list the realm in the
// store's successor:
//
//	func init(cur realm) { Relist(cross(cur)) }
//	func Relist(cur realm) { store.Register(cross(cur), ...) }
//
// The namespace owner can also list the realm with Claim, and the realm's
// next Register confirms it. Register takes its slug from a claim, which
// moves to one of its own (StoreClaimMoved), never its title from a
// confirmed listing. A slug is 3 to 40 of a-z, 0-9 and "-", with no "--".
// A core listing is frozen: its Register is refused (core_listing).
func Register(cur realm, slug, title, tagline, category string, palette int) {
	// The weekly update limit is updateCooldown.
	path := cur.Previous().PkgPath()
	kind := kindApp
	if existing := lookup(&byPath, path); existing != nil {
		kind = existing.Kind // a service claimed by its owner stays one
	}
	l := newListing(path, slug, title, tagline, category, palette, kind)
	l.Confirmed = true // deploying the realm proves it exists
	if reason := upsert(l, blockTime()); reason != "" {
		attrs := []string{"path", path, "reason", reason}
		if holder := holderOf(l, reason); holder != "" {
			attrs = append(attrs, "holder", holder) // what GovDAO would hide, if a squat
		}
		chain.Emit("StoreRegisterRejected", attrs...)
		return
	}
	emitListed("StoreRegistered", path)
}

// emitListed announces path's listing under its slug, once written. A
// hidden listing is announced under none: its hide released its slug, which
// another listing may hold by now.
func emitListed(event, path string, attrs ...string) {
	if l := lookup(&byPath, path); !l.hidden {
		chain.Emit(event, append([]string{"path", path, "slug", l.Slug}, attrs...)...)
	}
}

// holderOf returns the slug of the listing holding what l was refused for,
// its slug or its title, or "".
func holderOf(l *listing, reason string) string {
	switch reason {
	case reasonSlugTaken:
		return l.Slug
	case reasonTitleTaken:
		return titleHolder(l)
	}
	return ""
}

// Claim lists, or updates, path on behalf of its namespace owner: a /p/
// package (kind "package"), or a realm that cannot call Register (kind "app"
// or "service"). You must own the path's namespace. The store cannot check
// that a package exists, so a claimed listing is unconfirmed until it has 3
// stars that count toward its rank, or the realm's own Register confirms it.
// Until then a claimed package or service shows nowhere, a claimed app only
// at the end of its category, and it reserves no title: two claims may share
// one, which the first confirmed takes. Once confirmed, it reaches the
// shelves and the builders. A claim gives its slug up to a realm listing
// itself and moves to one of its own (StoreClaimMoved); claiming again takes
// the slug asked for, once free. Unlike Register, Claim must be called
// directly, and panics with the reason when it refuses; a core listing
// cannot be claimed (core_listing).
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
	l := newListing(path, slug, title, tagline, category, palette, kind)
	if reason := upsert(l, blockTime()); reason != "" {
		if holder := holderOf(l, reason); holder != "" {
			reason += " (holder: " + holder + ")"
		}
		panic("store: " + reason)
	}
	emitListed("StoreClaimed", path, "kind", kind)
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
	if l.core {
		panic("store: " + reasonCore) // frozen like its title: no image under the Core badge but the seed's
	}
	r := rich{Icon: icon, Cover: cover}
	if reason := validateRich(r); reason != "" {
		panic("store: " + reason)
	}
	l.submitRich(r, blockTime())
	if !l.hidden { // as emitListed: a hide released its slug
		chain.Emit("StoreRichSubmitted", "slug", l.Slug, "path", l.Path)
	}
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
