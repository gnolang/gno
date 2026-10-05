package store

import "strconv"

// The store front is a fixed sequence of grids (ADR §5): New and Top first
// and gno.land essentials last, always, with Trending, Recently updated and
// Rediscover between them when they have something to show. Most open onto a list of their own.

const (
	topStars = 3  // ranked stars an app needs to make Top
	gridPool = 16 // apps a section sends: gnoweb shows 8 once it sets aside what is shown above
)

// isTop reports an app with enough ranked stars for Top: below the floor, a
// young chain would crown an app with one star.
func isTop(l *listing) bool {
	return l.rankedStars.Size() >= topStars
}

// topApps returns at most n community apps of Top's head, the store front's
// and Block's: apps.stars is ordered by ranked stars, so the read stops at
// the first app below the floor (the list goes on below it; the front does
// not), and skips only the few core apps, which are in gno.land essentials.
func topApps(n int) []*listing {
	var out []*listing
	apps.stars.Iterate("", "", func(_ string, v any) bool {
		l := v.(*listing)
		if !isTop(l) {
			return true
		}
		if !l.core {
			out = append(out, l)
		}
		return len(out) >= n
	})
	return out
}

// newApps returns the n newest apps of New's head, the store front's and
// Block's: named namespaces only, like Rediscover, since a free g1…
// namespace earns no front-page slot (its apps stay in :latest and their
// category), and no core app, since those are in gno.land essentials.
func newApps(n int) []*listing {
	return window(namedApps, 0, n)
}

// homeShelves composes the store front at height h, over apps only.
func homeShelves(h int64) []shelf {
	newList, top := mustList("latest", h), mustList("top", h)
	return []shelf{
		{
			Title: newList.Title, More: newList, Pinned: true,
			// Its "See all", :latest, lists every app, so its count is
			// more than newApps draws from.
			Items: newApps(gridPool),
			Empty: "No new app yet. List yours from a registered namespace: it opens here.",
		},
		{
			Title: top.Title, Note: "Ranked by stars from accounts with a registered name.", More: top, Pinned: true,
			Items: topApps(gridPool),
			Empty: "No app has " + strconv.Itoa(topStars) + " stars from named accounts yet. Found one you like? Star it: that's how apps get here.",
		},
		listShelf(mustList("trending", h)),
		listShelf(mustList("updated", h)),
		{Title: "Rediscover", Note: "Older apps, a new set every day.", Items: rotation(h)},
		{Title: "gno.land essentials", Note: "Built by the core team. Good first stops.", Items: essentialApps(), Pinned: true, Empty: "No core app is listed."},
	}
}

// listShelf is the head of l, opening onto it, with its weekly counts. A
// computed list (Trending) is cut here, not computed shorter, so the count
// on its "See all" is that of its page. New and Top are not built here:
// their shelves filter what their lists hold (New's list, :latest, has
// every app; Top's goes on below the floor), so their counts are larger.
func listShelf(l list) shelf {
	s := shelf{Title: l.Title, More: l, Items: l.Items, Counts: l.Counts}
	switch {
	case l.Index != nil:
		s.Items = window(l.Index, 0, gridPool)
	case len(s.Items) > gridPool:
		s.Items, s.Counts = s.Items[:gridPool], s.Counts[:gridPool] // one count per item
	}
	return s
}

// essentials are the seed catalogue's listings, the core of gno.land, kept
// in seed order by the seed's init for gno.land essentials, which shows its
// apps, always. The community shelves (New, Top, Rediscover) leave them out
// by their core flag.
var essentials []*listing

// coreNamespaces are the core team's namespaces, kept out of the builder
// rankings, which they would otherwise top for good. The list is explicit:
// the seed also lists packages of personal namespaces (p/moul/md), whose
// owners are community builders like any other.
var coreNamespaces = []string{"gnoland", "gno", "demo", "gov", "sys", "nt", "gnops"}

func isCoreNamespace(ns string) bool {
	for _, c := range coreNamespaces {
		if c == ns {
			return true
		}
	}
	return false
}

func essentialApps() []*listing {
	var out []*listing
	for _, l := range essentials {
		if l.shownApp() { // a hidden core app, or one that became a service, is left out
			out = append(out, l)
		}
	}
	return out
}
