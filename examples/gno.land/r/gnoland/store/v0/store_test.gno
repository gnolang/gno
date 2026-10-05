package store

import (
	"chain/runtime"
	"strconv"
	"strings"
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
	"gno.land/r/sys/names"
	"gno.land/r/sys/users"
)

const testIcon = "ipfs://bafybeigdyrzt5sfp7udm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi"

// registerFrom cross-calls Register as the realm at path.
func registerFrom(cur realm, path, slug, title string) {
	registerIn(cur, path, slug, title, "utilities")
}

func registerIn(cur realm, path, slug, title, category string) {
	testing.SetRealm(testing.NewCodeRealm(path))
	Register(cross(cur), slug, title, "A tagline", category, 4)
}

// starFrom stars slug as each user, named by testutils.TestAddress, after
// giving the user a name the store has known for nameMaturity: only such
// stars rank.
func starFrom(cur realm, slug string, who ...string) {
	for _, u := range who {
		addr := nameUser(cur, u)
		if !voters.Has(addr.String()) {
			voters.Set(addr.String(), &voter{NameSeen: -nameMaturity})
		}
		testing.SetRealm(testing.NewUserRealm(addr))
		Star(cross(cur), slug)
	}
}

// nameUser gives the user u that name in r/sys/users and returns its address.
func nameUser(cur realm, u string) address {
	addr := testutils.TestAddress(u)
	if users.ResolveAddress(addr) == nil {
		testing.SetRealm(testing.NewCodeRealm("gno.land/r/sys/users/init"))
		if err := users.RegisterUserIgnoreCanonical(cross(cur), u, addr); err != nil {
			panic(err)
		}
	}
	return addr
}

// starAnon stars slug from addresses with no registered name.
func starAnon(cur realm, slug string, who ...string) {
	for _, u := range who {
		testing.SetRealm(testing.NewUserRealm(testutils.TestAddress(u)))
		Star(cross(cur), slug)
	}
}

func contains(t *testing.T, s, sub string) {
	t.Helper()
	uassert.True(t, strings.Contains(s, sub), sub+" not in "+s)
}

func TestRegisterFromRealm(cur realm, t *testing.T) {
	registerFrom(cur, "gno.land/r/acme/swap", "acme-swap", "Acme Swap")

	l := bySlugOrPanic("acme-swap")
	uassert.Equal(t, "gno.land/r/acme/swap", l.Path)
	uassert.Equal(t, kindApp, l.Kind)
	uassert.Equal(t, "acme", l.Namespace)
	contains(t, listingJSON("acme-swap"), `{"slug":"acme-swap","kind":"app","path":"gno.land/r/acme/swap","title":"Acme Swap","tagline":"A tagline","category":"utilities","palette":4,"stars":0,"ranked":0,"earned":false,"icon":"","cover":""}`)
	contains(t, Render("api/v1/category/utilities/1"), `"slug":"acme-swap"`)
	contains(t, Render("c/utilities"), "**[Acme Swap](/r/acme/swap)** — A tagline")
}

func TestRegisterRejectsCallersSilently(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(testutils.TestAddress("alice")))
	Register(cross(cur), "user-app", "User App", "Tagline", "utilities", 0)
	uassert.False(t, byPath.Has(""))

	// A `maketx run` script has an ephemeral /e/ path.
	const run = "gno.land/e/g1abc/run"
	testing.SetRealm(runtime.NewRealm(testutils.TestAddress("alice"), run))
	Register(cross(cur), "run-app", "Run App", "Tagline", "utilities", 0)
	uassert.False(t, byPath.Has(run))
}

func TestRegisterUpsert(cur realm, t *testing.T) {
	const path = "gno.land/r/upsert/app"
	registerFrom(cur, path, "upsert-app", "Upsert App")
	l := bySlugOrPanic("upsert-app")
	created := l.Created
	education := appCount("education")
	utilities := appCount("utilities")

	testing.SkipHeights(updateCooldown)
	testing.SetRealm(testing.NewCodeRealm(path))
	Register(cross(cur), "other-slug", "Upsert App v2", "New tagline", "education", 5)

	uassert.False(t, listings.Has("other-slug"), "the slug argument is ignored on update")
	uassert.True(t, l == bySlugOrPanic("upsert-app"))
	uassert.Equal(t, "Upsert App v2", l.Title)
	uassert.Equal(t, "education", l.Category)
	uassert.Equal(t, created, l.Created)
	uassert.Equal(t, created+updateCooldown, l.Updated)
	uassert.Equal(t, education+1, appCount("education"))
	uassert.Equal(t, utilities-1, appCount("utilities"))
	uassert.False(t, byTitle.Has(titleKey("Upsert App")), "the old title is released")
	uassert.Equal(t, `["upsert-app"]`, jsonSlugs(window(apps.updated, 0, 1)))
	uassert.True(t, activity.newestFirst()[0] == ev("updated", "upsert-app", 0))

	// An invalid update leaves the listing untouched.
	Register(cross(cur), "", "Bad\u202Etitle", "Tagline", "education", 0)
	uassert.Equal(t, "Upsert App v2", l.Title)
}

const spamPath = "gno.land/r/spam/app"

// registerSpam re-registers the spam-app listing with a tagline. testing.SetRealm
// applies to the calling frame, so it cannot live in a closure.
func registerSpam(cur realm, tagline string) {
	testing.SetRealm(testing.NewCodeRealm(spamPath))
	Register(cross(cur), "spam-app", "Spam App", tagline, "utilities", 4)
}

// TestUpdateSpam checks that re-registering buys no visibility: an
// unchanged listing is not bumped, and a real change bumps it at most once
// per updateCooldown, though the change itself is applied.
func TestUpdateSpam(cur realm, t *testing.T) {
	testing.SetHeight(900 * epochBlocks) // after every other update
	registerSpam(cur, "Version 1")
	l := bySlugOrPanic("spam-app")
	updatedShelf := func() string { return jsonSlugs(window(apps.updated, 0, 10)) }
	shelf, feed := updatedShelf(), activity

	for i := 0; i < 5; i++ {
		testing.SkipHeights(1)
		registerSpam(cur, "Version 1") // nothing changes
	}
	registerSpam(cur, "Version 2") // a real change, too soon
	uassert.Equal(t, "Version 2", l.Tagline, "the change is applied")
	uassert.Equal(t, l.Created, l.Updated, "but not announced")
	uassert.Equal(t, shelf, updatedShelf())
	uassert.True(t, feed == activity, "no activity entry")

	testing.SetRealm(testing.NewCodeRealm(spamPath))
	SubmitRich(cross(cur), testIcon, "")
	uassert.Equal(t, l.Created, l.Updated, "SubmitRich obeys the same cooldown")

	testing.SkipHeights(updateCooldown)
	registerSpam(cur, "Version 2") // unchanged: still no bump
	uassert.Equal(t, l.Created, l.Updated)
	registerSpam(cur, "Version 3")
	uassert.Equal(t, runtime.ChainHeight(), l.Updated)
	uassert.True(t, activity.newestFirst()[0] == ev("updated", "spam-app", 0))
	registerSpam(cur, "Version 4")
	uassert.Equal(t, "Version 4", l.Tagline)
	uassert.True(t, activity.newestFirst()[1].Kind != "updated" || activity.newestFirst()[1].Listing.Slug != "spam-app", "one announcement per cooldown")

	// Resubmitting the latest rich fields changes nothing.
	testing.SetRealm(testing.NewCodeRealm(spamPath))
	nextAt := l.nextAt
	testing.SkipHeights(1)
	SubmitRich(cross(cur), testIcon, "")
	uassert.Equal(t, nextAt, l.nextAt)
}

func TestRegisterValidation(cur realm, t *testing.T) {
	registerFrom(cur, "gno.land/r/valid/taken", "taken-slug", "Taken Title")

	cases := []struct {
		label, slug, title, tagline, category string
		palette                               int
		want                                  string
	}{
		{"short slug", "ab", "Title", "Tagline", "defi", 0, "bad_slug"},
		{"upper slug", "Bad-Slug", "Title", "Tagline", "defi", 0, "bad_slug"},
		{"taken slug", "taken-slug", "Title", "Tagline", "defi", 0, "slug_taken"},
		{"title case", "fresh-slug", "TAKEN title", "Tagline", "defi", 0, "title_taken"},
		{"bidi", "bidi-slug", "Swap\u202Ereverse", "Tagline", "defi", 0, "bad_title"},
		{"isolate", "iso-slug", "Swap\u2066x", "Tagline", "defi", 0, "bad_title"},
		{"filler", "fill-slug", "Gno\u3164Swap", "Tagline", "defi", 0, "bad_title"},
		{"braille blank", "braille-slug", "Title", "Tag\u2800line", "defi", 0, "bad_tagline"},
		{"zero width", "zw-slug", "Gno\u200BSwap", "Tagline", "defi", 0, "bad_title"},
		{"bom tagline", "bom-slug", "Title", "Tag\uFEFFline", "defi", 0, "bad_tagline"},
		{"control", "ctl-slug", "Ti\ntle", "Tagline", "defi", 0, "bad_title"},
		{"padded", "pad-slug", " Title", "Tagline", "defi", 0, "bad_title"},
		{"nbsp", "nbsp-slug", "Gno\u00A0Swap", "Tagline", "defi", 0, "bad_title"},
		{"ideographic space", "ideo-slug", "Title", "Tag\u3000line", "defi", 0, "bad_tagline"},
		{"greek omicron", "greek-slug", "Gn\u03bf.land Blog", "Tagline", "defi", 0, "bad_title"},
		{"cyrillic a", "cyr-slug", "Title", "Sw\u0430p tokens", "defi", 0, "bad_tagline"},
		{"armenian o", "arm-slug", "Gn\u0585 Swap", "Tagline", "defi", 0, "bad_title"},
		{"variation selector", "vs-slug", "Taken Title\uFE0F", "Tagline", "defi", 0, "bad_title"},
		{"grapheme joiner", "cgj-slug", "Gno\u034FSwap", "Tagline", "defi", 0, "bad_title"},
		{"combining accent", "acc-slug", "Cafe\u0301", "Tagline", "defi", 0, "bad_title"},
		{"mongolian selector", "fvs-slug", "Title", "Tag\u180Bline", "defi", 0, "bad_tagline"},
		{"fullwidth", "fw-slug", "\uFF27no Swap", "Tagline", "defi", 0, "bad_title"},
		{"math letters", "math-slug", "\U0001D406no Swap", "Tagline", "defi", 0, "bad_title"},
		{"double space", "dbl-slug", "Gno  Swap", "Tagline", "defi", 0, "bad_title"},
		{"capital i as l", "capi-slug", "Taken TitIe", "Tagline", "defi", 0, "title_taken"},
		{"digit as l", "digit-slug", "Taken Tit1e", "Tagline", "defi", 0, "title_taken"},
		{"roman numeral l", "roman-slug", "Gno.\u217Cand", "Tagline", "defi", 0, "bad_title"},
		{"ipa g", "ipa-slug", "\u0261no Swap", "Tagline", "defi", 0, "bad_title"},
		{"lisu letter", "lisu-slug", "\uA4D6no Swap", "Tagline", "defi", 0, "bad_title"},
		{"private use", "pua-slug", "Gno\uE000", "Tagline", "defi", 0, "bad_title"},
		{"unassigned", "cn-slug", "Gno\u0378", "Tagline", "defi", 0, "bad_title"},
		{"spaced title", "spaced-slug", "Taken-Title", "Tagline", "defi", 0, "title_taken"},
		{"cyrillic copy", "cyrcopy-slug", "\u0422\u0430ken Title", "Tagline", "defi", 0, "bad_title"},
		{"khmer mark", "khmer-slug", "Taken Title\u17B4", "Tagline", "defi", 0, "bad_title"},
		{"nukta on latin", "nukta-slug", "Gno\u093Cland", "Tagline", "defi", 0, "bad_title"},
		{"separators only", "dots-slug", "...", "Tagline", "defi", 0, "bad_title"},
		{"hyphen twin", "hyphen-slug", "Taken\u2010Title", "Tagline", "defi", 0, "title_taken"},
		{"long title", "long-slug", strings.Repeat("é", 41), "Tagline", "defi", 0, "bad_title"},
		{"category", "cat-slug", "Title", "Tagline", "casino", 0, "bad_category"},
		{"palette", "pal-slug", "Title", "Tagline", "defi", 12, "bad_palette"},
	}
	for _, tc := range cases {
		path := "gno.land/r/valid/" + strings.ReplaceAll(tc.label, " ", "")
		testing.SetRealm(testing.NewCodeRealm(path))
		Register(cross(cur), tc.slug, tc.title, tc.tagline, tc.category, tc.palette)
		uassert.False(t, byPath.Has(path), tc.label)
		l := newListing(path, tc.slug, tc.title, tc.tagline, tc.category, tc.palette, kindApp)
		uassert.Equal(t, tc.want, upsert(l, runtime.ChainHeight()), tc.label)
	}
}

func TestTitleRules(t *testing.T) {
	for _, ok := range []string{"Café", "Gno 東京", "Ελληνικά", "Ştiri", "Tiếng Việt", "नमस्ते Gno", "A1"} {
		uassert.True(t, validText(ok, maxTitle), ok)
	}
	for _, pair := range [][2]string{
		{"Gnoland", "Gno land"},
		{"Gnoland", "Gno-Land"},
		{"Gnoland", "Gnoiand"},
		{"Gnoland", "GnoIand"},
		{"ACME", "\u0410\u0421\u041C\u0415"}, // all Cyrillic
		{"Modern", "Modem"},
		{"Swap", "Svvap"},
	} {
		uassert.Equal(t, titleKey(pair[0]), titleKey(pair[1]), pair[1])
	}
	uassert.NotEqual(t, titleKey("Gnoland"), titleKey("Gnome"))
}

func TestNamespaceCap(cur realm, t *testing.T) {
	for i := 1; i <= maxPerNamespace+1; i++ {
		n := strconv.Itoa(i)
		registerFrom(cur, "gno.land/r/capped/app"+n, "capped-"+n, "Capped "+n)
	}
	uassert.Equal(t, maxPerNamespace, len(namespaceListings("capped")))
	uassert.False(t, listings.Has("capped-6"))
}

// TestClaimWithoutNames checks that, before r/sys/names is enabled, Claim
// accepts personal-address namespaces only. It runs before TestClaim, which
// enables names for good.
func TestClaimWithoutNames(cur realm, t *testing.T) {
	if names.IsEnabled() {
		t.Skip("r/sys/names is already enabled")
	}
	claim := func() {
		Claim(cross(cur), "gno.land/r/gnoland/blog", "defaced", "Defaced", "A tagline", "nft-art", 1, kindApp)
	}
	testing.SetRealm(testing.NewUserRealm(testutils.TestAddress("mallory")))
	uassert.AbortsContains(t, cur, "does not own the namespace", claim)
	uassert.True(t, ownsNamespace(testutils.TestAddress("mallory"), testutils.TestAddress("mallory").String()))
	uassert.Equal(t, "Gno.land Blog", bySlugOrPanic("blog").Title)
}

func TestClaim(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewUserRealm("g1skl80cuz8zq3lul9pgz5pc35l2pfzgxgfpsqkx")) // r/sys/names admin
	names.Enable(cross(cur))

	owner := testutils.TestAddress("claimer")
	ns := owner.String()
	path := "gno.land/p/" + ns + "/lib"
	claim := func() {
		Claim(cross(cur), path, "claimed-lib", "Claimed Lib", "A tagline", "dev-tools", 1, kindPackage)
	}

	testing.SetRealm(testing.NewUserRealm(testutils.TestAddress("mallory")))
	uassert.AbortsContains(t, cur, "does not own the namespace", claim)
	testing.SetRealm(runtime.NewRealm(owner, "gno.land/e/"+ns+"/run"))
	uassert.AbortsContains(t, cur, "direct user call", claim)

	testing.SetRealm(testing.NewUserRealm(owner))
	uassert.AbortsContains(t, cur, "store: bad_kind", func() {
		Claim(cross(cur), path, "claimed-lib", "Claimed Lib", "A tagline", "dev-tools", 1, kindApp)
	})
	claim()

	const slug = "claimed-lib"
	l := bySlugOrPanic(slug)
	uassert.Equal(t, kindPackage, l.Kind)
	uassert.Equal(t, ns, l.Namespace)
	newShelf := func() string {
		_, shelf, _ := strings.Cut(Render("api/v1/build"), `"title":"New to build with"`)
		shelf, _, _ = strings.Cut(shelf, "]")
		return shelf
	}
	uassert.False(t, l.Confirmed)
	uassert.False(t, strings.Contains(newShelf(), `"`+slug+`"`), "a claim alone buys no exposure")
	starAnon(cur, slug, "claimfan")
	starFrom(cur, slug, "claimfana", "claimfanb")
	uassert.False(t, l.Confirmed, "fewer than confirmStars ranked stars")
	starFrom(cur, slug, "claimfanc")
	uassert.True(t, l.Confirmed)
	contains(t, newShelf(), `"`+slug+`"`)
	uassert.False(t, strings.Contains(Render("api/v1/home"), slug), "packages stay off the store front")
}

// TestClaimedAppConfirmedByRegister checks that a realm listed by Claim
// reaches the shelves and the feed once its own Register confirms it.
func TestClaimedAppConfirmedByRegister(cur realm, t *testing.T) {
	owner := testutils.TestAddress("appowner")
	path := "gno.land/r/" + owner.String() + "/app"
	testing.SetRealm(testing.NewUserRealm(owner))
	Claim(cross(cur), path, "owned-app", "Owned App", "A tagline", "nft-art", 1, kindApp)
	l := bySlugOrPanic("owned-app")
	uassert.False(t, l.Confirmed)
	uassert.True(t, activity.newestFirst()[0].Listing.Slug != "owned-app", "no listed event yet")
	contains(t, Render("api/v1/category/nft-art/1"), `"slug":"owned-app"`)

	registerIn(cur, path, "ignored", "Owned App", "nft-art")
	uassert.True(t, l.Confirmed)
	uassert.True(t, activity.newestFirst()[0] == ev("listed", "owned-app", 0))
	uassert.True(t, apps.created.Has(descKey(l.Created, l.Slug)), "on the shelves")
}

// TestUnconfirmedAnnouncesNothing checks that a claim stays out of the feed
// and the pulse, even when updated.
func TestUnconfirmedAnnouncesNothing(cur realm, t *testing.T) {
	owner := testutils.TestAddress("ghostowner")
	path := "gno.land/r/" + owner.String() + "/ghost"
	feed, epoch := activity, runtime.ChainHeight()/epochBlocks
	listed := pulseListed.sum(epoch, 30)
	testing.SetRealm(testing.NewUserRealm(owner))
	Claim(cross(cur), path, "ghost-app", "Ghost App", "A tagline", "nft-art", 1, kindApp)
	testing.SkipHeights(updateCooldown)
	Claim(cross(cur), path, "ghost-app", "Ghost App", "Another tagline", "nft-art", 1, kindApp)
	uassert.Equal(t, "Another tagline", bySlugOrPanic("ghost-app").Tagline)
	uassert.True(t, feed == activity, "no event")
	uassert.Equal(t, listed, pulseListed.sum(epoch, 30))
	uassert.False(t, strings.Contains(Render("api/v1/home"), `"ghost-app"`))
}

// TestOwnStarDoesNotRank checks that a namespace owner's star on their own
// listing only adds to the displayed count.
func TestOwnStarDoesNotRank(cur realm, t *testing.T) {
	owner := nameUser(cur, "selfstar")
	voters.Set(owner.String(), &voter{NameSeen: -nameMaturity})
	path := "gno.land/r/" + owner.String() + "/self"
	testing.SetRealm(testing.NewUserRealm(owner))
	Claim(cross(cur), path, "self-app", "Self App", "A tagline", "nft-art", 1, kindApp)
	Star(cross(cur), "self-app")
	l := bySlugOrPanic("self-app")
	uassert.Equal(t, 1, l.stars.Size())
	uassert.Equal(t, 0, l.rankedStars.Size())
}

// TestRegisterKeepsServiceKind checks that a claimed service confirmed by
// its own Register stays a service.
func TestRegisterKeepsServiceKind(cur realm, t *testing.T) {
	owner := testutils.TestAddress("serviceowner")
	path := "gno.land/r/" + owner.String() + "/svc"
	testing.SetRealm(testing.NewUserRealm(owner))
	Claim(cross(cur), path, "owned-svc", "Owned Service", "A tagline", "infrastructure", 1, kindService)
	registerIn(cur, path, "ignored", "Owned Service", "infrastructure")
	l := bySlugOrPanic("owned-svc")
	uassert.True(t, l.Confirmed)
	uassert.Equal(t, kindService, l.Kind)
}

func TestPausedRegisterIsNoop(cur realm, t *testing.T) {
	paused = true
	registerFrom(cur, "gno.land/r/paused/app", "paused-app", "Paused App")
	uassert.False(t, listings.Has("paused-app"))

	paused = false
	registerFrom(cur, "gno.land/r/paused/app", "paused-app", "Paused App")
	uassert.True(t, listings.Has("paused-app"))
}

func TestHideUnhide(cur realm, t *testing.T) {
	const id = "coins"
	count := appCount("utilities")

	uassert.ErrorContains(t, hide(id, "dislike"), "unknown hide reason")
	uassert.NoError(t, hide(id, "broken"))
	uassert.False(t, listings.Has(id))
	for _, endpoint := range []string{"home", "build", "category/utilities/1"} {
		uassert.False(t, strings.Contains(Render("api/v1/"+endpoint), `"`+id+`"`), endpoint)
	}
	uassert.Equal(t, count-1, appCount("utilities"))
	testing.SetRealm(testing.NewUserRealm(testutils.TestAddress("fan")))
	uassert.AbortsContains(t, cur, "unknown listing", func() { Star(cross(cur), id) })

	// A hidden listing's changes are not announced.
	feed := activity
	testing.SkipHeights(updateCooldown)
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gnoland/coins"))
	SubmitRich(cross(cur), testIcon, "")
	uassert.True(t, feed == activity, "no event")

	uassert.ErrorContains(t, unhide("gno.land/r/gnoland/blog"), "not hidden")
	uassert.NoError(t, unhide("gno.land/r/gnoland/coins"))
	uassert.True(t, listings.Has(id))
	uassert.Equal(t, count, appCount("utilities"))
}

// TestHideReleasesTitleAndSlot checks that a hidden listing blocks no one,
// and that Unhide refuses when its title or slot was taken meanwhile.
func TestHideReleasesTitleAndSlot(cur realm, t *testing.T) {
	for i := 1; i <= maxPerNamespace; i++ {
		n := strconv.Itoa(i)
		registerFrom(cur, "gno.land/r/crowded/app"+n, "crowded-"+n, "Crowded "+n)
	}
	uassert.NoError(t, hide("crowded-1", "impersonation"))

	registerFrom(cur, "gno.land/r/crowded/newcomer", "crowded-new", "Crowded 1")
	uassert.True(t, listings.Has("crowded-new"), "the title and the slot were released")

	uassert.ErrorContains(t, unhide("gno.land/r/crowded/app1"), "store: title_taken")
	uassert.NoError(t, hide("crowded-new", "other"))
	registerFrom(cur, "gno.land/r/crowded/late", "crowded-late", "Crowded Late")
	uassert.ErrorContains(t, unhide("gno.land/r/crowded/app1"), "store: namespace_full")
}

// TestHideReleasesSlug checks that hiding a squatter frees its slug for the
// real app, and that Unhide then refuses.
func TestHideReleasesSlug(cur realm, t *testing.T) {
	registerFrom(cur, "gno.land/r/squatter/app", "brand-swap", "Brand Swap")
	uassert.NoError(t, hide("brand-swap", "impersonation"))
	uassert.ErrorContains(t, hide("brand-swap", "other"), "unknown listing")

	registerFrom(cur, "gno.land/r/brand/swap", "brand-swap", "Brand Swap")
	uassert.Equal(t, "gno.land/r/brand/swap", bySlugOrPanic("brand-swap").Path)
	for _, e := range activity.newestFirst() {
		uassert.False(t, e.Listing.Path == "gno.land/r/brand/swap" && e.Kind != "listed", "no event of the hidden listing follows the slug")
	}

	uassert.ErrorContains(t, unhide("gno.land/r/squatter/app"), "store: slug_taken")
}

func TestSubmitRich(cur realm, t *testing.T) {
	const path = "gno.land/r/rich/app"
	const cover = "ipfs://QmYwAPJzv5CZsnA625s3Xf2nemtYgPpHdWEz79ojWnPbdG"
	registerFrom(cur, path, "rich-app", "Rich App")

	testing.SetRealm(testing.NewCodeRealm("gno.land/r/rich/other"))
	uassert.AbortsContains(t, cur, "not a listed realm", func() {
		SubmitRich(cross(cur), testIcon, "")
	})

	testing.SetRealm(testing.NewCodeRealm(path))
	for _, bad := range [][2]string{
		{"https://example.com/logo.png", ""},
		{"data:image/png;base64,iVBORw0KGgo=", ""}, // inline images are refused: they bloat every list
		{"ipfs://short", ""},
		{"ipfs://" + strings.Repeat("a", 65), ""},
		{"ipfs://Qm<script>aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", ""},
		{"", "javascript:alert(1)"},
	} {
		uassert.AbortsContains(t, cur, "store: bad_", func() {
			SubmitRich(cross(cur), bad[0], bad[1])
		}, bad[0]+bad[1])
	}

	SubmitRich(cross(cur), testIcon, cover)
	l := bySlugOrPanic("rich-app")
	contains(t, listingJSON("rich-app"), `"icon":"","cover":""}`)
	_, pending := l.effective(runtime.ChainHeight())
	uassert.True(t, pending)

	testing.SkipHeights(richDelay)
	contains(t, listingJSON("rich-app"), `"icon":"`+testIcon+`","cover":"`+cover+`"}`)

	// A new submission replaces nothing visible until its own delay passes.
	SubmitRich(cross(cur), "", "")
	contains(t, listingJSON("rich-app"), `"icon":"`+testIcon+`"`)
	_, pending = l.effective(runtime.ChainHeight())
	uassert.True(t, pending)
}

func TestStarsAndEarned(cur realm, t *testing.T) {
	const id = "starred-app"
	registerFrom(cur, "gno.land/r/starred/app", "starred-app", "Starred App")
	l := bySlugOrPanic(id)

	testing.SetRealm(testing.NewCodeRealm("gno.land/r/starred/bot"))
	uassert.AbortsContains(t, cur, "direct user call", func() { Star(cross(cur), id) })

	fans := []string{"starfana", "starfanb", "starfanc", "starfand", "starfane"}
	starFrom(cur, id, fans[:earnStars-1]...)
	starAnon(cur, id, "anonymous")
	uassert.AbortsContains(t, cur, "already starred", func() { Star(cross(cur), id) })

	testing.SkipHeights(earnAge)
	uassert.False(t, l.earned(runtime.ChainHeight()), "the anonymous star must not count")
	starFrom(cur, id, fans[earnStars-1])
	uassert.False(t, l.earned(runtime.ChainHeight()), "a fresh star must not count")
	testing.SkipHeights(starMaturity)
	uassert.True(t, l.earned(runtime.ChainHeight()))

	testing.SetRealm(testing.NewUserRealm(testutils.TestAddress(fans[0])))
	Unstar(cross(cur), id)
	uassert.False(t, l.earned(runtime.ChainHeight()), "back below the threshold")
	uassert.AbortsContains(t, cur, "not starred", func() { Unstar(cross(cur), id) })
}

func TestStarRateLimit(cur realm, t *testing.T) {
	testing.SetHeight(500 * epochBlocks) // start of an epoch
	var slugs []string
	for i := 0; i <= maxStarsPerEpoch; i++ {
		ns := "ratens" + strconv.Itoa(i)
		registerFrom(cur, "gno.land/r/"+ns+"/app", ns, "Rate "+strconv.Itoa(i))
		slugs = append(slugs, ns)
	}
	testing.SetRealm(testing.NewUserRealm(testutils.TestAddress("bomber")))
	for _, id := range slugs[:maxStarsPerEpoch] {
		Star(cross(cur), id)
	}
	last := slugs[maxStarsPerEpoch]
	uassert.AbortsContains(t, cur, "star rate limit", func() { Star(cross(cur), last) })
	Unstar(cross(cur), slugs[0])
	uassert.AbortsContains(t, cur, "star rate limit", func() { Star(cross(cur), last) }, "unstar does not refund")

	testing.SkipHeights(epochBlocks)
	Star(cross(cur), last)
}

func TestStarsIndex(cur realm, t *testing.T) {
	registerFrom(cur, "gno.land/r/climber/app", "climber-app", "Climber")
	top := func() string { return jsonSlugs(window(apps.stars, 0, 3)) }
	before := top()
	var fans []string
	for i := 0; i < 12; i++ { // more than any other listing in this test run
		fans = append(fans, "climber"+strconv.Itoa(i))
	}
	starFrom(cur, "climber-app", fans...)
	contains(t, top(), `["climber-app"`)
	for _, fan := range fans {
		testing.SetRealm(testing.NewUserRealm(testutils.TestAddress(fan)))
		Unstar(cross(cur), "climber-app")
	}
	uassert.Equal(t, before, top())
}

// TestSybilStars checks that stars from addresses without a registered name
// are displayed but rank nothing.
func TestSybilStars(cur realm, t *testing.T) {
	testing.SetHeight(650 * epochBlocks)
	registerFrom(cur, "gno.land/r/sybil/app", "sybil-app", "Sybil Target")
	var bots []string
	for i := 0; i < milestones[0]; i++ {
		bots = append(bots, "sybilbot"+strconv.Itoa(i))
	}
	starAnon(cur, "sybil-app", bots...)

	l := bySlugOrPanic("sybil-app")
	contains(t, listingJSON("sybil-app"), `"stars":10`)
	uassert.Equal(t, 0, l.stars7d(runtime.ChainHeight()))
	uassert.False(t, strings.Contains(jsonSlugs(window(apps.stars, 0, 100)), "sybil-app"), "not in Top")
	uassert.False(t, strings.Contains(jsonSlugs(trending(runtime.ChainHeight(), 10).items), "sybil-app"), "not trending")
	uassert.True(t, activity.newestFirst()[0].Kind != "milestone", "no milestone")
	top, _ := topBuilders()
	for _, row := range top {
		uassert.False(t, row.Namespace == "sybil" && row.Stars > 0, "no builder stars")
	}
}

// TestNameMaturity checks that a name ranks stars only once the store has
// known it for nameMaturity.
func TestNameMaturity(cur realm, t *testing.T) {
	registerFrom(cur, "gno.land/r/mature/one", "mature-one", "Mature One")
	registerFrom(cur, "gno.land/r/mature/two", "mature-two", "Mature Two")
	addr := nameUser(cur, "newname")
	testing.SetRealm(testing.NewUserRealm(addr))
	Star(cross(cur), "mature-one")
	uassert.Equal(t, 0, bySlugOrPanic("mature-one").rankedStars.Size(), "a new name's star does not rank")

	testing.SkipHeights(nameMaturity)
	testing.SetRealm(testing.NewUserRealm(addr))
	Star(cross(cur), "mature-two")
	uassert.Equal(t, 1, bySlugOrPanic("mature-two").rankedStars.Size())
}

func TestStars7dAndTrending(cur realm, t *testing.T) {
	testing.SetHeight(600 * epochBlocks)
	const id = "weekly-app"
	registerFrom(cur, "gno.land/r/weekly/app", "weekly-app", "Weekly")
	l := bySlugOrPanic(id)

	starFrom(cur, id, "weeklya", "weeklyb", "weeklyc")
	testing.SkipHeights(epochBlocks)
	starFrom(cur, id, "weeklyd")
	uassert.Equal(t, 4, l.stars7d(runtime.ChainHeight()))
	contains(t, Render("api/v1/home"), `{"title":"Trending this week","slugs":["`+id+`"`)

	testing.SkipHeights(6 * epochBlocks) // the first epoch leaves the 7-epoch window
	uassert.Equal(t, 1, l.stars7d(runtime.ChainHeight()))
	uassert.Equal(t, 4, l.stars.Size())
	contains(t, Render("api/v1/home"), `{"title":"Trending this week","slugs":[]`)
}

// TestTrendingIgnoresStaleScores checks that apps that trended long ago,
// more than trendingScan of them, never hide an app trending now.
func TestTrendingIgnoresStaleScores(cur realm, t *testing.T) {
	testing.SetHeight(800 * epochBlocks) // heights only grow until TestTrendingIsExact
	for i := 0; i <= trendingScan; i++ {
		ns := "stale" + strconv.Itoa(i)
		registerFrom(cur, "gno.land/r/"+ns+"/app", ns, "Stale "+strconv.Itoa(i))
		starFrom(cur, ns, "stalefana", "stalefanb", "stalefanc")
		if (i+1)%(maxStarsPerEpoch/3) == 0 {
			testing.SkipHeights(epochBlocks) // the fans' star budget
		}
	}
	testing.SkipHeights(30 * epochBlocks)
	registerFrom(cur, "gno.land/r/fresh/app", "fresh-app", "Fresh App")
	starFrom(cur, "fresh-app", "freshfana", "freshfanb")
	top := trending(runtime.ChainHeight(), 10).items
	uassert.Equal(t, `["fresh-app"]`, jsonSlugs(top))
}

// TestTrendingIsExact checks that an app with more stars this week trends
// ahead of apps starred more recently.
func TestTrendingIsExact(cur realm, t *testing.T) {
	testing.SetHeight(960 * epochBlocks) // before TestBuilders, which needs the newest namespaces
	registerFrom(cur, "gno.land/r/hot/app", "hot-app", "Hot App")
	starFrom(cur, "hot-app", "hota", "hotb", "hotc", "hotd", "hote")
	testing.SkipHeights(epochBlocks)
	const n = 10
	for i := 0; i < n+2; i++ {
		ns := "warm" + strconv.Itoa(i)
		registerFrom(cur, "gno.land/r/"+ns+"/app", ns, "Warm "+strconv.Itoa(i))
		starFrom(cur, ns, ns+"a", ns+"b")
	}
	top := trending(runtime.ChainHeight(), n).items
	uassert.Equal(t, n, len(top))
	uassert.Equal(t, "hot-app", top[0].Slug)
}

func TestPulse(cur realm, t *testing.T) {
	testing.SetHeight(700 * epochBlocks) // earlier events are out of every window
	registerFrom(cur, "gno.land/r/pulse/one", "pulse-one", "Pulse One")
	registerFrom(cur, "gno.land/r/pulse/two", "pulse-two", "Pulse Two")
	starFrom(cur, "pulse-one", "pulsea", "pulseb", "pulsec")
	contains(t, Render("api/v1/home"), `"pulse":{"listed_7d":2,"stars_7d":3,"listed_30d":2,"stars_30d":3}`)

	testing.SkipHeights(updateCooldown) // the first epoch leaves the 7d window
	contains(t, Render("api/v1/home"), `"pulse":{"listed_7d":0,"stars_7d":0,"listed_30d":2,"stars_30d":3}`)
}

func TestActivity(cur realm, t *testing.T) {
	const id = "famous-app"
	registerFrom(cur, "gno.land/r/famous/app", "famous-app", "Famous")
	uassert.True(t, activity.newestFirst()[0] == ev("listed", id, 0))

	var fans []string
	for i := 0; i < milestones[0]; i++ {
		fans = append(fans, "famous"+strconv.Itoa(i))
	}
	starFrom(cur, id, fans...)
	uassert.True(t, activity.newestFirst()[0] == ev("milestone", id, milestones[0]))
	uassert.Equal(t, "listed", activity.newestFirst()[1].Kind)
	uassert.Equal(t, activitySize, len(activity.newestFirst()))

	// Falling below and climbing back does not announce the milestone twice.
	testing.SetRealm(testing.NewUserRealm(testutils.TestAddress(fans[0])))
	Unstar(cross(cur), id)
	starFrom(cur, id, "famouslate")
	uassert.Equal(t, "listed", activity.newestFirst()[1].Kind)

	home := Render("api/v1/home")
	contains(t, home, `"activity":[{"kind":"milestone","slug":"`+id+`","stars":10}`)
	_, refs, _ := strings.Cut(home, `"listings":`)
	contains(t, refs, `"slug":"`+id+`"`)
}

func TestBuilders(cur realm, t *testing.T) {
	testing.SetHeight(1000 * epochBlocks) // after every other listing
	registerFrom(cur, "gno.land/r/builderns/one", "builderns-one", "Builder One")
	testing.SkipHeights(1)
	registerFrom(cur, "gno.land/r/builderns/two", "builderns-two", "Builder Two")
	starFrom(cur, "builderns-two", "buildera", "builderb", "builderc")

	top, fresh := topBuilders()
	uassert.True(t, *fresh[0] == builder{"builderns", 2, 3, fresh[0].First})
	for i := 1; i < len(top); i++ {
		uassert.True(t, top[i-1].rank() >= top[i].rank(), "top builders by stars, then listings")
	}

	uassert.NoError(t, hide("builderns-two", "other"))
	_, fresh = topBuilders()
	uassert.True(t, *fresh[0] == builder{"builderns", 1, 0, fresh[0].First})
	contains(t, Render("api/v1/build"), `"builders":{"top":[{"namespace":`)
}

func TestRotation(t *testing.T) {
	const h = 10000 * epochBlocks // every app is old by then
	old := namedApps.Size()
	first := rotation(h)
	uassert.True(t, len(first) > 0 && len(first) <= rediscoverSize)
	uassert.Equal(t, jsonSlugs(first), jsonSlugs(rotation(h+1)), "same epoch, same window")
	uassert.NotEqual(t, jsonSlugs(first), jsonSlugs(rotation(h+epochBlocks)), "next epoch, next window")

	// Every app of a named namespace but the core ones is in the first half
	// of a window within ceil(old/rediscoverStep) epochs; address
	// namespaces and core apps never are.
	shown := map[string]bool{}
	for e := int64(0); e < int64((old+rediscoverStep-1)/rediscoverStep); e++ {
		w := rotation(h + e*epochBlocks)
		for i, l := range w {
			uassert.False(t, address(l.Namespace).IsValid(), l.Slug)
			uassert.False(t, l.core, l.Slug)
			if i < rediscoverStep {
				shown[l.Slug] = true
			}
		}
	}
	named := 0
	apps.created.Iterate("", "", func(_ string, v any) bool {
		if l := v.(*listing); !address(l.Namespace).IsValid() && !l.core {
			named++
		}
		return false
	})
	uassert.Equal(t, named, len(shown))
}

func TestHomeJSON(t *testing.T) {
	home := Render("api/v1/home")
	for _, want := range []string{
		`{"version":1,"height":`,
		`,"time":`,
		`"pulse":{"listed_7d":`,
		`"categories":[{"key":"defi","label":"DeFi","count":`,
		`{"key":"dao-governance","label":"DAO \u0026 Governance","count":`,
		`{"title":"Top","note":"Ranked by stars from accounts with a registered name.","pinned":true,"empty":"No app has 3 stars from named accounts yet. Found one you like? Star it: that's how apps get here.","slugs":["`,
		`"more":[{"key":"top","title":"Top","count":`,
		`{"title":"New","pinned":true,"empty":"No new app yet. List yours from a registered namespace: it opens here.","slugs":["`,
		`"more":[{"key":"latest","title":"New","count":`,
		`{"title":"Trending this week","slugs":[`,
		`"more":[{"key":"trending","title":"Trending this week","count":`,
		`"more":[{"key":"updated","title":"Recently updated","count":`,
		`"ranked":`,
		`{"title":"Rediscover","note":"Older apps, a new set every day.","slugs":[`,
		`{"title":"Recently updated","slugs":[`,
		`"activity":[{"kind":`,
		`"listings":[{"slug":`,
	} {
		contains(t, home, want)
	}
	uassert.False(t, strings.Contains(home, `"users"`), "services stay off the store front")
	uassert.Equal(t, notFound, Render("api/v1/nope"))
	uassert.Equal(t, notFound, Render("api/v1/app/blog"), "apps link to their own realm page")
	uassert.Equal(t, notFound, Render("api/v1/category/defi/0"))
	uassert.Equal(t, notFound, Render("api/v1/category/defi"))
	category := Render("api/v1/category/defi/1")
	contains(t, category, `{"version":1,"category":{"key":"defi","label":"DeFi","count":1},"page":1`)
	contains(t, category, `"page":1,"pages":1,"listings":[{"slug":`)
	contains(t, Render("api/v1/category/games/1"), `"page":1,"pages":1,"listings":[]`)
}

func TestJSONEscaping(cur realm, t *testing.T) {
	var b strings.Builder
	writeString(&b, "<\"\\>&\x01\u2028é")
	uassert.Equal(t, `"\u003c\"\\\u003e\u0026\u0001\u2028é"`, b.String())

	registerFrom(cur, "gno.land/r/escape/app", "escape-app", `A<"\>B`)
	contains(t, listingJSON("escape-app"), `"title":"A\u003c\"\\\u003eB"`)
	var item strings.Builder
	writeItems(&item, []*listing{bySlugOrPanic("escape-app")})
	contains(t, item.String(), `[A\<"\\\>B](/r/escape/app)`)
}

func jsonSlugs(ls []*listing) string {
	var b strings.Builder
	writeSlugs(&b, ls)
	return b.String()
}

// listingJSON is a listing's object as the API lists it.
func listingJSON(slug string) string {
	var b strings.Builder
	writeListing(&b, bySlugOrPanic(slug), runtime.ChainHeight())
	return b.String()
}

// ev is the event of kind for the listing with this slug.
func ev(kind, slug string, stars int) event {
	return event{kind, bySlugOrPanic(slug), stars}
}

func TestLists(t *testing.T) {
	latest := Render("api/v1/list/latest/1")
	contains(t, latest, `{"version":1,"list":{"key":"latest","title":"New","count":`)
	contains(t, latest, `"page":1,"pages":`)
	uassert.Equal(t, notFound, Render("api/v1/list/latest/0"))
	uassert.Equal(t, notFound, Render("api/v1/list/latest/999"))
	uassert.Equal(t, notFound, Render("api/v1/list/nope/1"))
	contains(t, Render("api/v1/list/top/1"), `"list":{"key":"top","title":"Top"`)
	trend := Render("api/v1/list/trending/1")
	contains(t, trend, `"list":{"key":"trending","title":"Trending this week"`)
	contains(t, trend, `"page":1,"pages":1,`)
	uassert.Equal(t, notFound, Render("api/v1/list/trending/2"), "a computed list is one page")
	contains(t, Render("api/v1/list/updated/1"), `"list":{"key":"updated","title":"Recently updated"`)

	md := Render("latest?page=2")
	contains(t, md, "# New")
	contains(t, md, "[Previous](/r/gnoland/store/v0:latest?page=1)")
	contains(t, Render(""), "[See all](/r/gnoland/store/v0:latest)")
	contains(t, Render(""), "## gno.land essentials")
}

func TestBuilderRankFits(t *testing.T) {
	uassert.True(t, maxPerNamespace < 16, "rank keeps listings in 4 bits")
}

// TestTopFloor checks that the front's Top holds only community apps with
// topStars ranked stars, while its list goes on below.
func TestTopFloor(t *testing.T) {
	for _, s := range homeShelves(runtime.ChainHeight()) {
		if s.Title != "Top" {
			continue
		}
		for _, l := range s.Items {
			uassert.True(t, isTop(l), l.Slug+" is below the floor")
			uassert.False(t, l.core, l.Slug+" is a core app")
		}
		uassert.True(t, s.More.count() >= len(s.Items))
	}
}

func TestCategoryMarkdownPages(cur realm, t *testing.T) {
	for i := 0; i < pageSize+1; i++ {
		ns, n := "pagens"+strconv.Itoa(i/maxPerNamespace), strconv.Itoa(i)
		registerIn(cur, "gno.land/r/"+ns+"/app"+n, "paged-"+n, "Paged "+n, "defi")
	}
	pages := strconv.Itoa(appCount("defi")/pageSize + 1)
	second := Render("c/defi?page=2")
	contains(t, second, "Page 2 of "+pages)
	contains(t, second, "[Previous](/r/gnoland/store/v0:c/defi?page=1)")
	uassert.False(t, strings.Contains(second, "Paged 24"), "the newest app is on page 1")
	first := Render("c/defi")
	contains(t, first, "**[Paged 24](/r/pagens4/app24)**")
	contains(t, first, "[Next](/r/gnoland/store/v0:c/defi?page=2)")
	contains(t, Render("c/defi?page=99"), "Page "+pages+" of "+pages)
}

// TestFrontSections checks the fixed order, and that New leaves the core
// apps to gno.land essentials, which is always drawn, and address
// namespaces to its list.
func TestFrontSections(t *testing.T) {
	var titles []string
	for _, s := range homeShelves(runtime.ChainHeight()) {
		titles = append(titles, s.Title)
		switch s.Title {
		case "New":
			uassert.True(t, len(s.Items) > 0, "core apps at the head of the index do not stop New")
			for _, l := range s.Items {
				uassert.False(t, l.core, l.Slug+" is a core app")
				uassert.False(t, address(l.Namespace).IsValid(), l.Slug+" is in an address namespace")
			}
		case "gno.land essentials":
			uassert.True(t, len(s.Items) > 0)
			uassert.True(t, s.Pinned)
		}
	}
	uassert.Equal(t, "New,Top,Trending this week,Recently updated,Rediscover,gno.land essentials", strings.Join(titles, ","))
}

// TestProposals checks that a proposal names an existing listing and a
// known reason. Execution by GovDAO is covered by the integration test
// store_govdao.txtar.
func TestProposals(cur realm, t *testing.T) {
	uassert.AbortsContains(t, cur, "unknown hide reason", func() { ProposeHide(cross(cur), "blog", "dislike") })
	uassert.AbortsContains(t, cur, "unknown listing", func() { ProposeHide(cross(cur), "nope", "other") })
	uassert.AbortsContains(t, cur, "not hidden", func() { ProposeUnhide(cross(cur), "gno.land/r/gnoland/blog") })
	ProposeHide(cross(cur), "blog", "other")
	ProposePause(cross(cur))
	ProposeUnpause(cross(cur))
}

// TestPick checks a GovDAO pick: an app only, for a bounded time, ended by
// a hide, and served first.
func TestPick(cur realm, t *testing.T) {
	h := runtime.ChainHeight()
	uassert.ErrorContains(t, setPick("nope", 7, h), "unknown listing")
	uassert.ErrorContains(t, setPick("avl", 7, h), "only an app")
	uassert.ErrorContains(t, setPick("blog", 0, h), "1 to 30 days")
	uassert.ErrorContains(t, setPick("blog", maxPickDays+1, h), "1 to 30 days")
	uassert.AbortsContains(t, cur, "1 to 30 days", func() { ProposeSpotlight(cross(cur), "blog", 31) })

	uassert.NoError(t, setPick("boards", 7, h))
	uassert.True(t, picked(h) == bySlugOrPanic("boards"))
	contains(t, Render("api/v1/home"), `"pick":"boards"`)
	contains(t, Render(""), "## Picked by GovDAO")
	uassert.True(t, picked(h+7*blocksPerDay) == nil, "the pick ends")

	uassert.NoError(t, hide("boards", "other"))
	uassert.True(t, picked(h) == nil, "a hidden app is no longer picked")
	uassert.NoError(t, unhide("gno.land/r/gnoland/boards2/v0"))
	uassert.True(t, picked(h) == nil, "nor once unhidden: the hide ended the pick")
	ProposeSpotlight(cross(cur), "boards", 7)
}

// TestPickEndsWithTheShownApp checks that a pick ends for good when its app
// stops being a shown app, and that it is bound to the title it was picked
// under.
func TestPickEndsWithTheShownApp(cur realm, t *testing.T) {
	owner := testutils.TestAddress("heroowner")
	path := "gno.land/r/" + owner.String() + "/hero"
	registerFrom(cur, path, "hero-app", "Hero App")
	h := runtime.ChainHeight()

	// A kind change app -> service -> app.
	uassert.NoError(t, setPick("hero-app", 7, h))
	testing.SetRealm(testing.NewUserRealm(owner))
	Claim(cross(cur), path, "hero-app", "Hero App", "A tagline", "utilities", 4, kindService)
	uassert.True(t, picked(h) == nil, "a service is not picked")
	Claim(cross(cur), path, "hero-app", "Hero App", "A tagline", "utilities", 4, kindApp)
	uassert.True(t, picked(h) == nil, "nor once an app again")

	// A rename, even undone.
	uassert.NoError(t, setPick("hero-app", 7, h))
	uassert.True(t, picked(h) == bySlugOrPanic("hero-app"))
	registerFrom(cur, path, "hero-app", "Hero App Renamed")
	uassert.True(t, picked(h) == nil, "a renamed app is not the one picked")
	registerFrom(cur, path, "hero-app", "Hero App")
	uassert.True(t, picked(h) == nil, "nor once renamed back")

	// A proposal is bound to the title it was made for.
	uassert.ErrorContains(t, pickBound("hero-app", path, titleKey("Other Title"), 7, h), "changed its title")
	uassert.NoError(t, pickBound("hero-app", path, titleKey("Hero App"), 7, h))
	uassert.True(t, picked(h) == bySlugOrPanic("hero-app"))
}

// TestNewSkipsAddressNamespaces checks that an app of a g1… namespace, even
// the newest, is not on the front's New but is on its list, :latest.
func TestNewSkipsAddressNamespaces(cur realm, t *testing.T) {
	testing.SetHeight(2000 * epochBlocks) // after every other listing
	path := "gno.land/r/" + testutils.TestAddress("newanon").String() + "/app"
	registerFrom(cur, path, "anon-new", "Anon New")
	testing.SkipHeights(1)
	registerFrom(cur, "gno.land/r/newnamed/app", "named-new", "Named New")

	front := homeShelves(runtime.ChainHeight())[0]
	uassert.Equal(t, "New", front.Title)
	uassert.Equal(t, "named-new", front.Items[0].Slug)
	uassert.False(t, strings.Contains(jsonSlugs(front.Items), `"anon-new"`))
	latest, _ := front.More.page(1)
	uassert.Equal(t, `["named-new","anon-new"`, jsonSlugs(latest)[:len(`["named-new","anon-new"`)])
}

// TestSeedNamespacesNotBuilders checks that the core namespaces stay out of
// both builder rankings, even with a community listing among them, while a
// personal namespace the seed lists a package of (moul) is a builder.
func TestSeedNamespacesNotBuilders(cur realm, t *testing.T) {
	testing.SetHeight(3000 * epochBlocks) // after every other listing
	registerFrom(cur, "gno.land/r/gnoland/extra", "gnoland-extra", "Gnoland Extra")
	starFrom(cur, "gnoland-extra", "seedfana", "seedfanb", "seedfanc")
	testing.SkipHeights(1)
	registerFrom(cur, "gno.land/r/community/app", "community-app", "Community App")

	top, fresh := topBuilders()
	uassert.True(t, len(top) > 0 && len(fresh) > 0)
	uassert.Equal(t, "community", fresh[0].Namespace)
	for _, row := range append(top, fresh...) {
		uassert.False(t, isCoreNamespace(row.Namespace), row.Namespace+" is a core namespace")
	}
	for _, ns := range []string{"gnoland", "demo", "gov", "gnops", "sys", "nt"} {
		uassert.True(t, isCoreNamespace(ns), ns)
		uassert.False(t, builders.Has(ns), ns+" has a builder")
	}

	uassert.False(t, isCoreNamespace("moul"))
	registerFrom(cur, "gno.land/r/moul/storeapp", "moul-app", "Moul App")
	b, _ := builders.Get("moul").(*builder)
	uassert.True(t, b != nil && b.Listings == 2, "moul's seeded package and app count")
}

// TestProposalBoundToPath checks that a hide or a pick proposed for one
// listing never executes against another that took its slug since.
func TestProposalBoundToPath(cur realm, t *testing.T) {
	const victim, squatter = "gno.land/r/boundvictim/app", "gno.land/r/boundsquat/app"
	registerFrom(cur, victim, "bound-app", "Bound App")
	ProposeHide(cross(cur), "bound-app", "other")   // checks, then binds to victim
	ProposeSpotlight(cross(cur), "bound-app", 7)    // likewise, and to its title
	uassert.NoError(t, hide("bound-app", "broken")) // another proposal frees the slug
	registerFrom(cur, squatter, "bound-app", "Bound App")

	uassert.ErrorContains(t, hideBound("bound-app", victim, "other"), "another listing")
	uassert.ErrorContains(t, pickBound("bound-app", victim, titleKey("Bound App"), 7, runtime.ChainHeight()), "another listing")
	uassert.False(t, bySlugOrPanic("bound-app").hidden)
	uassert.True(t, picked(runtime.ChainHeight()) == nil || picked(runtime.ChainHeight()).Path != squatter)
	uassert.NoError(t, pickBound("bound-app", squatter, titleKey("Bound App"), 1, runtime.ChainHeight()))
	uassert.NoError(t, hideBound("bound-app", squatter, "other"))
}

// TestPickNeedsConfirmed checks that a claimed app no one confirmed cannot
// be picked: picked() would not show it.
func TestPickNeedsConfirmed(cur realm, t *testing.T) {
	owner := testutils.TestAddress("pickowner")
	testing.SetRealm(testing.NewUserRealm(owner))
	Claim(cross(cur), "gno.land/r/"+owner.String()+"/app", "unconfirmed-pick", "Unconfirmed Pick", "A tagline", "nft-art", 1, kindApp)
	uassert.ErrorContains(t, setPick("unconfirmed-pick", 7, runtime.ChainHeight()), "only a confirmed app")
	uassert.AbortsContains(t, cur, "only a confirmed app", func() { ProposeSpotlight(cross(cur), "unconfirmed-pick", 7) })
}

// TestBlock checks the embeddable lists: the front's rules, n clamped, the
// titles sanitized, and nothing for an unknown list.
func TestBlock(cur realm, t *testing.T) {
	testing.SetHeight(4000 * epochBlocks) // after every other listing
	registerFrom(cur, "gno.land/r/blockns/app", "block-app", "Block *App* [x](y)")
	latest := Block("latest", 0)
	uassert.Equal(t, 1, strings.Count(latest, "- "), "n clamped to 1")
	contains(t, latest, "**[Block \\*App\\* \\[x\\]\\(y\\)](/r/blockns/app)**")
	contains(t, latest, "[See all on Explore](/r/gnoland/store/v0:latest)")
	uassert.True(t, strings.Count(Block("latest", 99), "\n- ") <= maxBlock)
	uassert.False(t, strings.Contains(Block("latest", 8), "/r/gnoland/blog"), "no core app")
	contains(t, Block("top", 8), "(/r/gnoland/store/v0:top)")
	contains(t, Block("trending", 8), "(/r/gnoland/store/v0:trending)")
	uassert.Equal(t, "", Block("updated", 8))
	uassert.Equal(t, "", Block("nope", 8))
}

// TestBadge checks that a badge shows the store's own ranked count, and
// nothing for an unlisted or hidden app.
func TestBadge(cur realm, t *testing.T) {
	registerFrom(cur, "gno.land/r/badgens/app", "badge-app", "Badge App")
	starFrom(cur, "badge-app", "badgefana", "badgefanb")
	starAnon(cur, "badge-app", "badgeanon")
	uassert.Equal(t, "★ 2 on Explore · [Star](/r/gnoland/store/v0$help&func=Star&slug=badge-app)", Badge("gno.land/r/badgens/app"))
	uassert.Equal(t, "", Badge("gno.land/r/badgens/nope"))
	uassert.Equal(t, "", Badge("gno.land/p/nt/avl/v0"), "not an app")
	uassert.NoError(t, hide("badge-app", "other"))
	uassert.Equal(t, "", Badge("gno.land/r/badgens/app"))
}
