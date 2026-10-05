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
	"chain/runtime"

	"gno.land/p/nt/avl/v0"
	"gno.land/r/sys/names"
	"gno.land/r/sys/users"
)

const (
	blocksPerDay = 17280 // 5 s blocks

	newWindow   = 14 * blocksPerDay // an app is new for Rediscover and "New to build with"
	epochBlocks = blocksPerDay      // rotation and counter period
	richDelay   = blocksPerDay      // publication delay of rich fields (T1, T8)
	// Launch values for a young chain (ADR §2): earned reputation only gates
	// owner assets and the Spotlight, so it may be reachable early. They are
	// constants: a realm upgrade raises them as the chain grows.
	earnAge      = 14 * blocksPerDay // minimum listing age for earned reputation
	starMaturity = 3 * blocksPerDay  // a star counts towards reputation after this
	earnStars    = 5                 // qualifying stars needed for earned reputation
	// nameMaturity is how long the store must have known an address's name
	// before its stars rank: names are free on a young chain, so a name made
	// for one campaign has to wait (T8).
	nameMaturity = 3 * blocksPerDay
	// confirmStars ranked stars confirm a claimed listing (T11).
	confirmStars = 3
	// updateCooldown is the least time between two "updated" announcements of
	// a listing, so that re-registering in a loop buys no visibility.
	updateCooldown = 7 * epochBlocks

	// No global cap: reads are bounded by shelf sizes, trendingScan and
	// pagination, and storage deposits price every write.
	maxPerNamespace  = 5  // visible apps, services and packages together
	maxStarsPerEpoch = 20 // per address, against star bombing
)

// voters holds what the store keeps per starring address.
var voters avl.Tree // address -> voter, a value: a pointer would store one more object per address

type voter struct {
	Epoch    int64 // of the last star
	Spent    int   // stars spent in Epoch
	NameSeen int64 // height the store first saw the address with a registered name, 0 if never
}

// Register lists the calling realm as an app. It must be cross-called by the
// app realm itself, typically from its init():
//
//	store.Register(cross(cur), "acme-swap", "Acme Swap", "Swap tokens in one click", "defi", 3)
//
// Calling Register again, as an upgraded realm's init does, updates the
// listing: the original slug and creation height are kept and the slug
// argument is ignored. A real change is announced as an update at most once
// a week. Register never panics, so it cannot break the caller's init; a
// refused call emits StoreRegisterRejected with the path and a short reason.
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
	reason := upsert(l, runtime.ChainHeight())
	if reason != "" {
		chain.Emit("StoreRegisterRejected", "path", path, "reason", reason)
	}
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
	if reason := upsert(newListing(path, slug, title, tagline, category, palette, kind), runtime.ChainHeight()); reason != "" {
		panic("store: " + reason)
	}
}

// ownsNamespace reports whether addr may list paths of namespace ns.
// r/sys/names authorises everyone until a chain enables it (dev chains,
// tests), so the store then accepts personal-address namespaces only: one
// user can never edit another's listing.
func ownsNamespace(addr address, ns string) bool {
	if !names.IsEnabled() {
		return addr.String() == ns
	}
	return names.IsAuthorizedAddressForNamespace(addr, ns)
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
	l := byPathOrPanic(cur.Previous().PkgPath())
	r := rich{Icon: icon, Cover: cover}
	if reason := validateRich(r); reason != "" {
		panic("store: " + reason)
	}
	l.submitRich(r, runtime.ChainHeight())
}

// Star an app you like. One star per app, up to 20 new stars a day. Your
// star counts toward an app's rank once your address has a registered @name
// the store has seen for 3 days; until then it shows in the app's total.
// Stars on apps in your own namespace never rank.
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
	if l.starred(addr) {
		panic("store: already starred")
	}
	h := runtime.ChainHeight()
	named := users.ResolveAddress(caller) != nil
	nameSeen := spendStar(addr, h, named)
	ranked := named && h-nameSeen >= nameMaturity && !ownsNamespace(caller, l.Namespace)
	l.addStar(addr, star{Height: h, Ranked: ranked})
}

// Unstar removes your star from an app. It does not give back that day's
// star: the 20 a day count the stars you add.
func Unstar(cur realm, slug string) {
	l, caller := starTarget(cur, slug)
	if !l.starred(caller.String()) {
		panic("store: not starred")
	}
	l.removeStar(caller.String(), runtime.ChainHeight())
}

func starTarget(cur realm, slug string) (*listing, address) {
	caller := cur.Previous()
	if !caller.IsUserCall() {
		panic("store: stars must come from a direct user call")
	}
	return bySlugOrPanic(slug), caller.Address()
}

// spendStar charges one star to addr's budget for the epoch of h, records
// h as the height the store first saw addr named if it is the first, and
// returns that height (0 if never).
func spendStar(addr string, h int64, named bool) int64 {
	v, _ := voters.Get(addr).(voter)
	if epoch := h / epochBlocks; v.Epoch != epoch {
		v.Epoch, v.Spent = epoch, 0
	}
	if v.Spent >= maxStarsPerEpoch {
		panic("store: star rate limit")
	}
	v.Spent++
	if named && v.NameSeen == 0 {
		v.NameSeen = h
	}
	voters.Set(addr, v)
	return v.NameSeen
}
