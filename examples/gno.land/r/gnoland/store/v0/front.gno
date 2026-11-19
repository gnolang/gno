package store

// The store front is a fixed set of sections (ADR §5): New, Top, Trending,
// Recently updated, Rediscover and gno.land essentials. Most open onto a
// list of their own. gnoweb decides which it draws, under which name, and
// may ask more of an app than the floors here (topStars, minTrending).

const (
	topStars = 3  // ranked stars an app needs to make Top
	gridPool = 16 // apps a section sends: gnoweb shows 8 once it sets aside what is shown above
)

// isTop reports an app with enough ranked stars for Top: below the floor, a
// young chain would crown an app with one star.
func isTop(l *listing) bool {
	return l.rankedCount() >= topStars
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
	return window(&namedApps, 0, n)
}

// homeShelves composes the store front at time now, over apps only.
func homeShelves(now int64) []shelf {
	newList, top, updated := mustList("latest", now), mustList("top", now), mustList("updated", now)
	return []shelf{
		// The "See all" of New and of Recently updated, :latest and
		// :updated, list every app: more than their shelves draw from.
		{Key: "new", Title: newList.Title, More: newList, Items: newApps(gridPool)},
		{Key: "top", Title: top.Title, More: top, Items: topApps(gridPool)},
		listShelf(mustList("trending", now)),
		{Key: "updated", Title: updated.Title, More: updated, Items: window(&namedUpdated, 0, gridPool)},
		{Key: "rediscover", Title: "Rediscover", Items: rotation(now)},
		{Key: "core", Title: "gno.land essentials", Items: essentialApps()},
	}
}

// listShelf is the head of l, opening onto it. A computed list (Trending)
// is cut here, not computed shorter, so the count on its "See all" is that
// of its page. New and Top are not built here: their shelves filter what
// their lists hold (New's list, :latest, has every app; Top's goes on below
// the floor), so their counts are larger.
func listShelf(l list) shelf {
	s := shelf{Key: l.Key, Title: l.Title, More: l, Items: l.Items}
	switch {
	case l.Index != nil:
		s.Items = window(l.Index, 0, gridPool)
	case len(s.Items) > gridPool:
		s.Items = s.Items[:gridPool]
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
