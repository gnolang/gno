package store

import (
	"strconv"
	"strings"
	"testing"
	"unicode/utf8"
)

// The checks below rebuild every index from the listings themselves and
// compare: the store can never be redeployed, so an index left out of step
// by any sequence of writes would stay so. TestInvariants runs them on what
// the suite leaves, runRandom after each of its random writes.

type nsTotals struct {
	n, stars int
	first    int64
}

// checkInvariants rebuilds every index from byPath, compares, and returns
// what does not hold.
func checkInvariants(now int64) (bad []string) {
	fail := func(m string) { bad = append(bad, m) }
	visible, titles := 0, 0
	hiddenCores := 0
	slugOwner, titleOwner := map[string]string{}, map[string]string{}
	catAll, catConf := map[string]int{}, map[string]int{}
	nsAgg, nsVisible := map[string]*nsTotals{}, map[string]int{}
	var nCreated, nUpdated, nStars [2]int
	nNamed, nNamedUpd, nTrend := 0, 0, 0
	trendPerEpoch := map[int64]int{}
	var trendy []*listing

	byPath.Iterate("", "", func(path string, v any) bool {
		l := v.(*listing)
		p := path + " (" + l.Slug + "): "
		if path != l.Path || l.Namespace != namespaceOf(l.Path) || l.Namespace == "" {
			fail(p + "path/namespace")
		}
		if r := validateListing(l); r != "" {
			fail(p + "stored invalid: " + r)
		}
		if !validSlug(l.Slug) && !isMovedSlug(l) {
			fail(p + "slug neither asked for nor a moved claim's")
		}
		vis := !l.hidden
		if vis != (lookup(&listings, l.Slug) == l) {
			fail(p + "listed by slug iff visible")
		}
		if vis {
			visible++
			if o, dup := slugOwner[l.Slug]; dup {
				fail(p + "slug shared with " + o)
			}
			slugOwner[l.Slug] = path
			if !l.core { // a core listing is unhidden whatever its namespace holds
				nsVisible[l.Namespace]++
			}
		}
		key := titleKey(l.Title)
		holds := vis && l.Confirmed
		got, _ := byTitle.Get(key).(string)
		if holds {
			titles++
			if got != l.Slug {
				fail(p + "confirmed and visible but byTitle[" + key + "]=" + got)
			}
			if o, dup := titleOwner[key]; dup {
				fail(p + "title shared by two confirmed listings, other " + o)
			}
			titleOwner[key] = path
		} else if vis && got == l.Slug {
			fail(p + "a claim holds a title")
		}
		if l.core && l.hidden {
			hiddenCores++
		}
		if l.core {
			if !l.Confirmed || !l.selfListed {
				fail(p + "core not confirmed")
			}
			if l.hidden && (listings.Has(l.Slug) || byTitle.Has(key)) {
				fail(p + "hidden core lost its slug or title")
			}
		}
		if l.Confirmed && !l.selfListed && l.rankedCount() < confirmStars {
			fail(p + "star-confirmed with " + strconv.Itoa(l.rankedCount()) + " ranked stars")
		}
		if l.indexed() && !l.announced {
			fail(p + "indexed but never announced")
		}

		// category
		inCat := lookup(categoryMembers(l.Category), l.catKey()) == l
		if (vis && l.Kind == kindApp) != inCat {
			fail(p + "in its category iff a visible app")
		}
		if vis && l.Kind == kindApp {
			catAll[l.Category]++
			if l.Confirmed {
				catConf[l.Category]++
			}
		}

		// shelves
		k, ix := 0, &apps
		if l.Kind != kindApp {
			k, ix = 1, &toolbox
		}
		ckey := descKey(l.Created, l.Slug)
		if l.indexed() != (lookup(&ix.created, ckey) == l) {
			fail(p + "created shelf")
		}
		ukey := descKey(l.Updated, l.Slug)
		wantUpd := l.indexed() && l.Updated != l.Created
		if wantUpd != (lookup(&ix.updated, ukey) == l) {
			fail(p + "updated shelf")
		}
		wantStars := l.indexed() && l.rankedCount() > 0
		if wantStars != (lookup(&ix.stars, descKey(int64(l.rankedCount()), l.Slug)) == l) {
			fail(p + "stars shelf")
		}
		wantNamed := l.indexed() && l.namedApp() && !l.core
		if wantNamed != (lookup(&namedApps, ckey) == l) {
			fail(p + "namedApps")
		}
		wantNamedUpd := wantUpd && l.namedApp()
		if wantNamedUpd != (lookup(&namedUpdated, ukey) == l) {
			fail(p + "namedUpdated")
		}
		wantTrend := l.shownApp() && l.rankedCount() > 0 && l.trendScore >= minTrending
		if wantTrend != (lookup(&appTrend, trendKey(l)) == l) {
			fail(p + "appTrend")
		}
		if l.indexed() {
			nCreated[k]++
			if wantUpd {
				nUpdated[k]++
			}
			if wantStars {
				nStars[k]++
			}
		}
		if wantNamed {
			nNamed++
		}
		if wantNamedUpd {
			nNamedUpd++
		}
		if wantTrend {
			nTrend++
			trendPerEpoch[l.trendEpoch]++
		}

		for _, m := range checkStars(l, now) {
			fail(p + m)
		}
		if l.shownApp() && l.stars7d(now) >= minTrending {
			trendy = append(trendy, l)
		}

		// builders
		if l.indexed() && !isCoreNamespace(l.Namespace) {
			a := nsAgg[l.Namespace]
			if a == nil {
				a = &nsTotals{first: l.Created}
				nsAgg[l.Namespace] = a
			}
			a.n++
			a.stars += l.rankedCount()
			if l.Created < a.first {
				a.first = l.Created
			}
		}
		return false
	})

	eq := func(what string, got, want int) {
		if got != want {
			fail(what + ": " + strconv.Itoa(got) + ", want " + strconv.Itoa(want))
		}
	}
	eq("listings.Size", listings.Size(), visible)
	eq("byTitle.Size", byTitle.Size(), titles)
	eq("hiddenCore", hiddenCore, hiddenCores)
	eq("apps.created", apps.created.Size(), nCreated[0])
	eq("apps.updated", apps.updated.Size(), nUpdated[0])
	eq("apps.stars", apps.stars.Size(), nStars[0])
	eq("toolbox.created", toolbox.created.Size(), nCreated[1])
	eq("toolbox.updated", toolbox.updated.Size(), nUpdated[1])
	eq("toolbox.stars", toolbox.stars.Size(), nStars[1])
	eq("namedApps", namedApps.Size(), nNamed)
	eq("namedUpdated", namedUpdated.Size(), nNamedUpd)
	eq("appTrend", appTrend.Size(), nTrend)
	for ns, n := range nsVisible {
		if !isCoreNamespace(ns) && n > maxPerNamespace {
			fail("namespace " + ns + " holds " + strconv.Itoa(n) + " visible listings besides the core ones")
		}
	}
	cats := 0
	for _, c := range categories {
		eq("appCount "+c.Key, appCount(c.Key), catAll[c.Key])
		eq("confirmedCount "+c.Key, confirmedCount(c.Key), catConf[c.Key])
		if byCategory.Has(c.Key) {
			cats++
		}
		seenClaim := false
		categoryMembers(c.Key).Iterate("", "", func(_ string, v any) bool {
			l := v.(*listing)
			if !l.Confirmed {
				seenClaim = true
			} else if seenClaim {
				fail("category " + c.Key + ": confirmed " + l.Slug + " after a claim")
			}
			return false
		})
	}
	eq("byCategory keys", byCategory.Size(), cats)

	eq("builders", builders.Size(), len(nsAgg))
	eq("buildersByRank", buildersByRank.Size(), len(nsAgg))
	eq("buildersByFirst", buildersByFirst.Size(), len(nsAgg))
	for ns, a := range nsAgg {
		b, _ := builders.Get(ns).(*builder)
		if b == nil {
			fail("builder " + ns + " missing")
			continue
		}
		if b.Listings != a.n || b.Stars != a.stars || b.First != a.first {
			fail("builder " + ns + " totals")
		}
		if r, _ := buildersByRank.Get(descKey(b.rank(), ns)).(*builder); r != b {
			fail("builder " + ns + " rank key")
		}
		if r, _ := buildersByFirst.Get(descKey(b.First, ns)).(*builder); r != b {
			fail("builder " + ns + " first key")
		}
	}

	// trending is exact while no epoch holds more than its share of the budget.
	exact := true
	for _, n := range trendPerEpoch {
		if n > trendingScan/trendEpochs {
			exact = false
		}
	}
	got := trending(now, pageSize)
	if exact {
		for i := 1; i < len(trendy); i++ {
			for j := i; j > 0; j-- {
				a, b := trendy[j].stars7d(now), trendy[j-1].stars7d(now)
				if !(a > b || a == b && trendy[j].Slug < trendy[j-1].Slug) {
					break
				}
				trendy[j], trendy[j-1] = trendy[j-1], trendy[j]
			}
		}
		if len(trendy) > pageSize {
			trendy = trendy[:pageSize]
		}
		if jsonSlugs(trendy) != jsonSlugs(got.items) {
			fail("trending " + jsonSlugs(got.items) + ", want " + jsonSlugs(trendy))
		}
	}
	if l := picked(now); l != nil && !l.shownApp() {
		fail("picked a listing that is not shown")
	}
	return bad
}

// jsonChecker is a strict JSON validator.
type jsonChecker struct {
	s string
	i int
}

func validJSON(s string) bool {
	p := &jsonChecker{s: s}
	return utf8.ValidString(s) && p.value() && p.i == len(s)
}

func (p *jsonChecker) eat(c byte) bool {
	if p.i < len(p.s) && p.s[p.i] == c {
		p.i++
		return true
	}
	return false
}

func (p *jsonChecker) value() bool {
	if p.i >= len(p.s) {
		return false
	}
	switch c := p.s[p.i]; {
	case c == '{':
		p.i++
		if p.eat('}') {
			return true
		}
		for {
			if !p.str() || !p.eat(':') || !p.value() {
				return false
			}
			if p.eat('}') {
				return true
			}
			if !p.eat(',') {
				return false
			}
		}
	case c == '[':
		p.i++
		if p.eat(']') {
			return true
		}
		for {
			if !p.value() {
				return false
			}
			if p.eat(']') {
				return true
			}
			if !p.eat(',') {
				return false
			}
		}
	case c == '"':
		return p.str()
	case c == 't':
		return p.lit("true")
	case c == 'f':
		return p.lit("false")
	case c == 'n':
		return p.lit("null")
	}
	start := p.i
	p.eat('-')
	for p.i < len(p.s) && p.s[p.i] >= '0' && p.s[p.i] <= '9' {
		p.i++
	}
	return p.i > start && p.s[p.i-1] != '-'
}

func (p *jsonChecker) lit(w string) bool {
	if strings.HasPrefix(p.s[p.i:], w) {
		p.i += len(w)
		return true
	}
	return false
}

func (p *jsonChecker) str() bool {
	if !p.eat('"') {
		return false
	}
	for p.i < len(p.s) {
		c := p.s[p.i]
		p.i++
		switch {
		case c == '"':
			return true
		case c < 0x20:
			return false
		case c == '\\':
			if p.i >= len(p.s) {
				return false
			}
			e := p.s[p.i]
			p.i++
			if e == 'u' {
				for k := 0; k < 4; k++ {
					if p.i >= len(p.s) || !strings.ContainsRune("0123456789abcdefABCDEF", rune(p.s[p.i])) {
						return false
					}
					p.i++
				}
			} else if !strings.ContainsRune(`"\/bfnrt`, rune(e)) {
				return false
			}
		}
	}
	return false
}

func apiEndpoints() []string {
	out := []string{"api/v1/home", "api/v1/build"}
	for _, c := range categories {
		out = append(out, "api/v1/category/"+c.Key+"/1")
	}
	for _, k := range []string{"top", "latest", "updated", "trending"} {
		out = append(out, "api/v1/list/"+k+"/1")
	}
	return out
}

func checkJSON() []string {
	var bad []string
	for _, e := range apiEndpoints() {
		if out := Render(e); !validJSON(out) {
			bad = append(bad, e+" is not valid JSON: "+out)
		}
	}
	return bad
}

func tryStar(cur realm, slug string, addr address) string {
	testing.SetRealm(testing.NewUserRealm(addr))
	return abortReason(revive(func() { Star(cross(cur), slug) }))
}

func tryUnstar(cur realm, slug string, addr address) string {
	testing.SetRealm(testing.NewUserRealm(addr))
	return abortReason(revive(func() { Unstar(cross(cur), slug) }))
}

func tryClaim(cur realm, owner address, path, slug, title, category, kind string) string {
	testing.SetRealm(testing.NewUserRealm(owner))
	return abortReason(revive(func() { Claim(cross(cur), path, slug, title, "A tagline", category, 1, kind) }))
}

func abortReason(r any) string {
	switch v := r.(type) {
	case nil:
		return ""
	case string:
		return v
	case error:
		return v.Error()
	}
	return "abort (non-string)"
}

// isMovedSlug reports a slug the store gave a claim it moved (movedSlug):
// the one kind no listing may ask for.
func isMovedSlug(l *listing) bool {
	return strings.Contains(l.Slug, movedMark) && len(l.Slug) <= maxSlug
}

func checkStars(l *listing, now int64) []string {
	var bad []string
	epoch := now / epochLength
	if l.book == nil {
		if l.trendScore != 0 {
			bad = append(bad, "trendScore without stars")
		}
		return bad
	}
	ranked, in7, missing := 0, 0, 0
	book := l.book
	book.all.Iterate("", "", func(addr string, v any) bool {
		p := v.(int64)
		at := p / 2
		if p%2 == 1 {
			ranked++
			if !book.ranked.Has(descKey(at, addr)) {
				missing++
			}
			if e := at / epochLength; e <= epoch && epoch-e < trendEpochs {
				in7++
			}
		}
		return false
	})
	if missing > 0 {
		bad = append(bad, "ranked star missing from ranked tree")
	}
	if ranked != l.rankedCount() {
		bad = append(bad, "ranked tree size "+strconv.Itoa(l.rankedCount())+" != "+strconv.Itoa(ranked))
	}
	if got := l.stars7d(now); got != in7 {
		bad = append(bad, "stars7d "+strconv.Itoa(got)+" != held ranked stars of the window "+strconv.Itoa(in7))
	}
	if l.trendScore < l.stars7d(now) || l.trendEpoch > epoch {
		bad = append(bad, "trendScore is not an upper bound")
	}
	if at := l.stars7d(l.trendEpoch * epochLength); l.trendScore != at {
		bad = append(bad, "trendScore "+strconv.Itoa(l.trendScore)+" != stars7d at trendEpoch "+strconv.Itoa(at))
	}
	return bad
}
