package store

import (
	"strconv"
	"strings"
	"testing"

	"gno.land/p/nt/testutils/v0"
)

type lcg struct{ s uint64 }

func (r *lcg) n(k int) int {
	r.s = r.s*6364136223846793005 + 1442695040888963407
	return int((r.s >> 33) % uint64(k))
}

func coreSnapshot() string {
	var b strings.Builder
	for _, l := range essentials {
		b.WriteString(l.Path + "|" + l.Slug + "|" + l.Kind + "|" + l.Title + "|" + l.Tagline + "|" + l.Category + "|" + strconv.Itoa(l.Palette) + "|" + l.icon + "|" + l.cover + "|")
		if l.next != nil {
			b.WriteString("next")
		}
		b.WriteString(strconv.FormatBool(l.Confirmed) + strconv.FormatBool(l.core) + "\n")
	}
	return b.String()
}

func slugsByPath() map[string]string {
	m := map[string]string{}
	byPath.Iterate("", "", func(path string, v any) bool {
		l := v.(*listing)
		h := "v"
		if l.hidden {
			h = "h"
		}
		m[path] = h + l.Slug
		return false
	})
	return m
}

// checkEvery is how many random writes pass between two full checks: one
// rebuilds every index of a store the whole suite has filled.
const checkEvery = 50

// runRandom applies steps random writes from seed (listings, claims, stars,
// hides, unhides, picks, images, time passing). After each it checks that
// the write ended as a write may, and that no listing was deleted; every
// checkEvery writes, every invariant, that no star was lost, that a slug
// only moved for a claim, once, and that no core listing changed. tag keeps
// the names of two runs apart.
func runRandom(cur realm, t *testing.T, tag string, seed uint64, steps int, startEpoch int64) {
	enableNames(cur)
	setEpoch(startEpoch)
	rng := &lcg{s: seed}

	owners := map[string]address{}
	var nss []string
	for i := 0; i < 4; i++ {
		name := "zr" + tag + "n" + strconv.Itoa(i)
		owners[name] = nameUser(cur, name)
		nss = append(nss, name)
	}
	for i := 0; i < 4; i++ {
		a := testutils.TestAddress("zr" + tag + "addr" + strconv.Itoa(i))
		owners[a.String()] = a
		nss = append(nss, a.String())
	}
	var paths []string
	for _, ns := range nss {
		for i := 0; i < 5; i++ {
			paths = append(paths, "gno.land/r/"+ns+"/a"+strconv.Itoa(i))
		}
		for i := 0; i < 2; i++ {
			paths = append(paths, "gno.land/p/"+ns+"/l"+strconv.Itoa(i))
		}
	}
	var slugs []string
	for i := 0; i < 10; i++ {
		slugs = append(slugs, "z"+tag+"-s"+strconv.Itoa(i))
	}
	// The longest slugs, and those a moved claim's would once have run into.
	long := "z" + tag + "-" + strings.Repeat("x", maxSlug-len(tag)-2)
	slugs = append(slugs, long, long[:maxSlug-2]+"-2", long[:maxSlug-2]+"-3", long[:maxSlug-1], long[:maxSlug-3]+"-10")
	slugs = append(slugs, "z"+tag+"-s0-2", "z"+tag+"-s0-3", "z"+tag+"-s1-2", "boards", "coins", "markdown")
	var titles []string
	for i := 0; i < 60; i++ {
		w := string(rune('a'+i%26)) + "o" + string(rune('a'+(i*7)%26)) + string(rune('a'+i/26))
		titles = append(titles, "Zed "+tag+" "+w)
		if i < 8 {
			titles = append(titles, "ZED-"+strings.ToUpper(tag)+"-"+strings.ToUpper(w), "Zed "+tag+" "+strings.ReplaceAll(w, "o", "0"))
		}
	}
	titles = append(titles, "Boards", "Coin Balances", "B0ards", "Markdown Helpers")
	cats := []string{"games", "education", "nft-art"}
	var ranked, anon, late []address
	for i := 0; i < 12; i++ {
		ranked = append(ranked, rankedFan(cur, "zr"+tag+"fan"+strconv.Itoa(i)))
	}
	for i := 0; i < 3; i++ {
		anon = append(anon, testutils.TestAddress("zr"+tag+"anon"+strconv.Itoa(i)))
		late = append(late, nameUser(cur, "zr"+tag+"late"+strconv.Itoa(i)))
	}

	model := map[string]map[string]bool{}                  // path -> starring addresses
	byPath.Iterate("", "", func(path string, v any) bool { // stars an earlier run left
		if l := v.(*listing); l.book != nil {
			fans := map[string]bool{}
			l.book.all.Iterate("", "", func(a string, _ any) bool { fans[a] = true; return false })
			model[path] = fans
		}
		return false
	})
	core := coreSnapshot()
	var log []string
	stats := map[string]int{}

	expected := []string{"slug_taken", "title_taken", "namespace_full", "core_listing", "bad_kind", "already starred",
		"star rate limit", "not starred", "unknown listing", "not a listed realm", "does not own the namespace"}
	refused := func(op, r string) {
		if r == "" {
			stats[op+" ok"]++
			return
		}
		for _, e := range expected {
			if strings.Contains(r, e) {
				stats[op+" "+e]++
				return
			}
		}
		t.Errorf("UNEXPECTED ABORT in %s: %s", op, r)
	}
	visibleAt := func(i int) *listing {
		_, v := listings.GetByIndex(i % listings.Size())
		return v.(*listing)
	}
	// claimOr returns a random visible unconfirmed claim half the time.
	claimOr := func(l *listing) *listing {
		if rng.n(2) == 0 {
			return l
		}
		var claims []*listing
		listings.Iterate("", "", func(_ string, v any) bool {
			if !v.(*listing).Confirmed {
				claims = append(claims, v.(*listing))
			}
			return false
		})
		if len(claims) == 0 {
			return l
		}
		return claims[rng.n(len(claims))]
	}

	before := slugsByPath()
	for step := 0; step < steps && !t.Failed(); step++ {
		size := byPath.Size()
		op := ""
		switch w := rng.n(100); {
		case w < 15: // Register
			path, slug, title, cat := paths[rng.n(len(paths))], slugs[rng.n(len(slugs))], titles[rng.n(len(titles))], cats[rng.n(len(cats))]
			if rng.n(12) == 0 {
				path = essentials[rng.n(len(essentials))].Path
			}
			if strings.Contains(path, "/p/") {
				continue
			}
			if c := claimOr(nil); c != nil && rng.n(2) == 0 {
				slug = c.Slug // a realm listing itself over a claim's slug
			}
			op = "Register(" + path + "," + slug + "," + title + "," + cat + ")"
			testing.SetRealm(testing.NewCodeRealm(path))
			if r := revive(func() { Register(cross(cur), slug, title, "A tagline", cat, rng.n(palettes)) }); r != nil {
				t.Errorf("Register PANICS: %v", r)
			}
			stats["register"]++
		case w < 30: // Claim
			path, slug, title, cat := paths[rng.n(len(paths))], slugs[rng.n(len(slugs))], titles[rng.n(len(titles))], cats[rng.n(len(cats))]
			kind := kindApp
			if strings.Contains(path, "/p/") {
				kind = kindPackage
			} else if rng.n(3) == 0 {
				kind = kindService
			}
			if rng.n(20) == 0 {
				kind = kindPackage
			}
			ns := namespaceOf(path)
			op = "Claim(" + path + "," + slug + "," + title + "," + cat + "," + kind + ")"
			refused("claim", tryClaim(cur, owners[ns], path, slug, title, cat, kind))
		case w < 55: // Star
			l := claimOr(visibleAt(rng.n(1 << 20)))
			var who address
			switch a := rng.n(20); {
			case a < 13:
				who = ranked[rng.n(len(ranked))]
			case a < 15:
				who = anon[rng.n(len(anon))]
			case a < 17:
				who = late[rng.n(len(late))]
			default:
				if o, ok := owners[l.Namespace]; ok {
					who = o
				} else {
					who = ranked[0]
				}
			}
			op = "Star(" + l.Slug + " " + l.Path + "," + who.String()[:8] + ")"
			r := tryStar(cur, l.Slug, who)
			refused("star", r)
			if r == "" {
				if model[l.Path] == nil {
					model[l.Path] = map[string]bool{}
				}
				model[l.Path][who.String()] = true
			}
		case w < 68: // Unstar
			l := visibleAt(rng.n(1 << 20))
			fans := model[l.Path]
			if len(fans) == 0 {
				continue
			}
			var all []string
			l.book.all.Iterate("", "", func(a string, _ any) bool { all = append(all, a); return false })
			who := all[rng.n(len(all))]
			op = "Unstar(" + l.Slug + " " + l.Path + "," + who[:8] + ")"
			r := tryUnstar(cur, l.Slug, address(who))
			refused("unstar", r)
			if r == "" {
				delete(fans, who)
			}
		case w < 75: // hide
			l := visibleAt(rng.n(1 << 20))
			if l.core && rng.n(3) > 0 {
				continue
			}
			op = "hidePath(" + l.Path + ")"
			if err := hidePath(l.Path, "other"); err != nil {
				t.Errorf("hide of a visible listing fails: %v", err)
			}
			if !l.hidden {
				t.Errorf("hidePath did not hide its listing")
			}
			stats["hide"]++
		case w < 83: // unhide
			var hidden []*listing
			byPath.Iterate("", "", func(_ string, v any) bool {
				if v.(*listing).hidden {
					hidden = append(hidden, v.(*listing))
				}
				return false
			})
			if len(hidden) == 0 {
				continue
			}
			l := hidden[rng.n(len(hidden))]
			op = "unhide(" + l.Path + ")"
			err := unhide(l.Path)
			if err != nil {
				stats["unhide "+err.Error()]++
				if l.core {
					t.Errorf("a hidden core listing cannot be unhidden: %v", err)
				}
			} else {
				stats["unhide ok"]++
			}
		case w < 93: // time
			d := []int64{1, 3600, day, day, 2 * day, 3 * day, 8 * day, 15 * day}[rng.n(8)]
			op = "skip(" + strconv.FormatInt(d, 10) + ")"
			skip(d)
		case w < 97: // SubmitRich
			path := paths[rng.n(len(paths))]
			if strings.Contains(path, "/p/") {
				continue
			}
			icon := []string{"", testIcon, testIcon[:len(testIcon)-1] + "j"}[rng.n(3)]
			op = "SubmitRich(" + path + ")"
			testing.SetRealm(testing.NewCodeRealm(path))
			refused("rich", abortReason(revive(func() { SubmitRich(cross(cur), icon, "") })))
		default: // pick
			l := visibleAt(rng.n(1 << 20))
			op = "setPick(" + l.Slug + ")"
			if err := setPick(l.Slug, 1+rng.n(maxPickDays), blockTime()); err == nil {
				stats["pick ok"]++
			}
		}
		log = append(log, strconv.Itoa(step)+" "+op)

		// No listing is ever deleted.
		if byPath.Size() < size {
			t.Errorf("a listing was deleted")
		}
		if (step+1)%checkEvery != 0 && step != steps-1 {
			continue
		}

		// No star is lost, and a slug moves only for a claim, once.
		after := slugsByPath()
		for path, was := range before {
			now, ok := after[path]
			if !ok {
				t.Errorf("%s was deleted", path)
			} else if now[1:] != was[1:] {
				stats["reslug"]++
				if l := lookup(&byPath, path); !isMovedSlug(l) || strings.Contains(was, movedMark) {
					t.Errorf("%s moved from %s to %s", path, was, now)
				}
			}
		}
		before = after
		for path, fans := range model {
			l := lookup(&byPath, path)
			if l == nil || l.starCount() != len(fans) {
				t.Errorf("%s lost stars: model %d", path, len(fans))
				continue
			}
			for a := range fans {
				if !l.starred(a) {
					t.Errorf("%s lost the star of %s", path, a)
				}
			}
		}
		if now := coreSnapshot(); now != core {
			t.Errorf("a core listing changed:\n%s\nwas\n%s", now, core)
		}
		for _, m := range checkInvariants(blockTime()) {
			t.Errorf("INVARIANT %s", m)
		}
	}
	if t.Failed() {
		from := len(log) - 14
		if from < 0 {
			from = 0
		}
		t.Logf("seed %d, last operations:\n%s", seed, strings.Join(log[from:], "\n"))
	}
	// A run that never did one of these checked less than it says.
	for _, did := range []string{"claim ok", "register", "star ok", "unstar ok", "hide", "unhide ok", "reslug"} {
		if stats[did] == 0 {
			t.Errorf("seed %d never did %q", seed, did)
		}
	}
}
