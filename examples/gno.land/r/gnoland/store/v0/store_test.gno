package store

import (
	"chain/runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
	"gno.land/r/sys/names"
	"gno.land/r/sys/users"
)

// origin is the epoch the tests' clock starts from: past the default test
// block time, so setEpoch only ever moves the clock forward.
const origin = 20000

// setEpoch moves the block time to the start of epoch origin+e.
func setEpoch(e int64) {
	setClock((origin + e) * epochLength)
}

// skip moves the block time d seconds forward.
func skip(d int64) {
	setClock(blockTime() + d)
}

func setClock(t int64) {
	ctx := testing.GetContext()
	ctx.Time = time.Unix(t, 0)
	testing.SetContext(ctx)
}

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
			voters.Set(addr.String(), voter{NameSeen: -nameMaturity})
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
	contains(t, listingJSON("acme-swap"), `{"slug":"acme-swap","kind":"app","path":"gno.land/r/acme/swap","title":"Acme Swap","tagline":"A tagline","category":"utilities","palette":4,"stars":0,"ranked":0,"fifth_ranked":0,"created":`+strconv.FormatInt(bySlugOrPanic("acme-swap").Created, 10)+`,"pending":false,"core":false,"icon":"","cover":""}`)
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

// Durations follow block time, not heights: a chain faster or slower than
// a block every 5 s still waits a real day for a day.
func TestDurationsAreBlockTime(cur realm, t *testing.T) {
	registerFrom(cur, "gno.land/r/clock/app", "clock-app", "Clock App")
	l := bySlugOrPanic("clock-app")

	testing.SetHeight(runtime.ChainHeight() + newWindow) // blocks, no time
	uassert.True(t, l.isNew(blockTime()), "heights alone must not age a listing")

	skip(newWindow)
	uassert.False(t, l.isNew(blockTime()), "newWindow of block time ages it")
}

func TestRegisterUpsert(cur realm, t *testing.T) {
	const path = "gno.land/r/upsert/app"
	registerFrom(cur, path, "upsert-app", "Upsert App")
	l := bySlugOrPanic("upsert-app")
	created := l.Created
	education := appCount("education")
	utilities := appCount("utilities")

	skip(updateCooldown)
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
	uassert.Equal(t, `["upsert-app"]`, jsonSlugs(window(&apps.updated, 0, 1)))
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
	setEpoch(900) // after every other update
	registerSpam(cur, "Version 1")
	l := bySlugOrPanic("spam-app")
	updatedShelf := func() string { return jsonSlugs(window(&apps.updated, 0, 10)) }
	shelf, feed := updatedShelf(), activity

	for i := 0; i < 5; i++ {
		skip(1)
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

	skip(updateCooldown)
	registerSpam(cur, "Version 2") // unchanged: still no bump
	uassert.Equal(t, l.Created, l.Updated)
	registerSpam(cur, "Version 3")
	uassert.Equal(t, blockTime(), l.Updated)
	uassert.True(t, activity.newestFirst()[0] == ev("updated", "spam-app", 0))
	registerSpam(cur, "Version 4")
	uassert.Equal(t, "Version 4", l.Tagline)
	uassert.True(t, activity.newestFirst()[1].Kind != "updated" || activity.newestFirst()[1].Listing.Slug != "spam-app", "one announcement per cooldown")

	// Resubmitting the latest rich fields changes nothing.
	testing.SetRealm(testing.NewCodeRealm(spamPath))
	nextAt := l.nextAt
	skip(1)
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
		uassert.Equal(t, tc.want, upsert(l, blockTime()), tc.label)
	}
}

func TestTitleRules(t *testing.T) {
	for _, ok := range []string{"Café", "Gno 東京", "Ελληνικά", "Ştiri", "Tiếng Việt", "नमस्ते Gno", "A1", "Rock ’n’ Roll"} {
		uassert.True(t, validTitle(ok), ok)
	}
	uassert.False(t, validTitle("\u03A0\u041E\u0420\u0422"), "Greek Pi among Cyrillic letters")
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

// Each copy the reviews found of a listed title is refused, or keys as the
// original does: none can list beside it.
func TestTitleLookalikes(t *testing.T) {
	for _, pair := range [][2]string{
		{"TOKEN", "ΤΟΚΕΝ"}, {"ZONE", "ΖΟΝΕ"}, {"MINT", "ΜΙΝΤ"}, {"KEY", "ΚΕΥ"}, // Greek capitals
		{"ACME", "ꓮꓚꓟꓰ"}, {"ACME", "ᎪᏟᎷᎬ"}, // Lisu, Cherokee
		{"Boards", "B၀ards"}, {"Gno.land Blog", "Gn०.land Blog"}, {"Gnoland Swap", "Gn০land Swap"}, // foreign zeros
		{"Gno.land Blog", "Gno.∣and Blog"}, {"Gnoland Swap", "Gno׀and Swap"}, // bars
		{"Gnoland Swap", "Gnoاand Swap"}, {"Gnoland Swap", "Gnסland Swap"}, // Arabic alef, Hebrew samekh
		{"Gnoland Pay", "GnoƖand Pay"}, {"Gnoland Pay", "Gnoland ƿay"}, // Latin Extended-B
		{"Zoom", "ℤℴℴℳ"}, {"CHN", "ℂℍℕ"}, // Letterlike Symbols
		{"Boards", "Bóards"}, {"Gno.land Blog", "Gno.land Blòg"}, {"Coin Balances", "Cöin Balances"}, {"GovDAO", "GovDAÖ"}, // precomposed accents
		{"FOX", "ϜΟΧ"}, {"MAP", "ϺΑΡ"}, {"YES", "ϒΕS"}, // archaic Greek capitals
		{"Gnoland", "Gno\u05D5\u05B9and"},                                                                          // Hebrew vav with a mark, drawn as an i
		{"Gno.land Blog", "Gno.ㅣand Blog"}, {"Gnoland", "Gno丨and"}, {"Gnoland", "Gnㅇland"}, {"Gnoland", "Gnهland"}, // Hangul, Han, Arabic
		{"AEON", "ÆON"}, {"OEIL", "ŒIL"}, // ligatures
		{"Gnoland", "Gnoﺍand"}, {"Gnoland", "Gno\uFB4Band"}, {"Gnoland", "Gnoￜand"}, {"Gnoland", "Gno\u1175and"}, // compatibility forms
		{"Gnoland", "Gnoأand"},                             // alef with hamza
		{"ПОРТ", "ΠΟΡΤ"}, {"ГОРА", "ΓΟΡΑ"}, {"ПОРТ", "ΠОРТ"}, // Greek and Cyrillic, alike or mixed
		{"カメラ", "力メラ"}, {"ヘルプ", "へルプ"}, // Han and hiragana drawn as katakana
		{"Bob's Swap", "Bob’s Swap"}, // a typographic apostrophe
	} {
		uassert.True(t, !validTitle(pair[1]) || titleKey(pair[1]) == titleKey(pair[0]), pair[1])
	}
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
	uassert.True(t, apps.created.Has(descKey(l.Created, l.Slug)), "on the shelves")
	uassert.True(t, activity.newestFirst()[0].Listing != l, "a g1… namespace is not announced")
}

// TestUnconfirmedAnnouncesNothing checks that a claim stays out of the feed
// and the pulse, even when updated.
func TestUnconfirmedAnnouncesNothing(cur realm, t *testing.T) {
	owner := testutils.TestAddress("ghostowner")
	path := "gno.land/r/" + owner.String() + "/ghost"
	feed, epoch := activity, blockTime()/epochLength
	listed := pulseListed.sum(epoch, 30)
	testing.SetRealm(testing.NewUserRealm(owner))
	Claim(cross(cur), path, "ghost-app", "Ghost App", "A tagline", "nft-art", 1, kindApp)
	skip(updateCooldown)
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
	voters.Set(owner.String(), voter{NameSeen: -nameMaturity})
	path := "gno.land/r/" + owner.String() + "/self"
	testing.SetRealm(testing.NewUserRealm(owner))
	Claim(cross(cur), path, "self-app", "Self App", "A tagline", "nft-art", 1, kindApp)
	Star(cross(cur), "self-app")
	l := bySlugOrPanic("self-app")
	uassert.Equal(t, 1, l.starCount())
	uassert.Equal(t, 0, l.rankedCount())
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
	skip(updateCooldown)
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
	_, pending := l.effective(blockTime())
	uassert.True(t, pending)

	skip(richDelay)
	contains(t, listingJSON("rich-app"), `"icon":"`+testIcon+`","cover":"`+cover+`"}`)

	// A new submission replaces nothing visible until its own delay passes.
	SubmitRich(cross(cur), "", "")
	contains(t, listingJSON("rich-app"), `"icon":"`+testIcon+`"`)
	_, pending = l.effective(blockTime())
	uassert.True(t, pending)
}

func TestStarsAndFifthRanked(cur realm, t *testing.T) {
	const id = "starred-app"
	registerFrom(cur, "gno.land/r/starred/app", "starred-app", "Starred App")
	l := bySlugOrPanic(id)

	testing.SetRealm(testing.NewCodeRealm("gno.land/r/starred/bot"))
	uassert.AbortsContains(t, cur, "direct user call", func() { Star(cross(cur), id) })

	fans := []string{"starfana", "starfanb", "starfanc", "starfand", "starfane"}
	starFrom(cur, id, fans[:earnStars-1]...)
	starAnon(cur, id, "anonymous")
	uassert.AbortsContains(t, cur, "already starred", func() { Star(cross(cur), id) })

	skip(day)
	uassert.Equal(t, int64(0), l.fifthRanked(), "the anonymous star must not count")
	starFrom(cur, id, fans[earnStars-1])
	fifth := blockTime()
	uassert.Equal(t, fifth, l.fifthRanked())
	skip(day)
	starFrom(cur, id, "starfanf")
	uassert.Equal(t, fifth, l.fifthRanked(), "a newer star leaves the fifth oldest")
	contains(t, listingJSON(id), `"fifth_ranked":`+strconv.FormatInt(fifth, 10))

	testing.SetRealm(testing.NewUserRealm(testutils.TestAddress(fans[0])))
	Unstar(cross(cur), id)
	uassert.True(t, l.fifthRanked() > fifth, "the fifth oldest is now a newer star")
	uassert.AbortsContains(t, cur, "not starred", func() { Unstar(cross(cur), id) })
}

func TestStarRateLimit(cur realm, t *testing.T) {
	setEpoch(500) // start of an epoch
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

	skip(epochLength)
	Star(cross(cur), last)
}

func TestStarsIndex(cur realm, t *testing.T) {
	registerFrom(cur, "gno.land/r/climber/app", "climber-app", "Climber")
	top := func() string { return jsonSlugs(window(&apps.stars, 0, 3)) }
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
	setEpoch(650)
	registerFrom(cur, "gno.land/r/sybil/app", "sybil-app", "Sybil Target")
	var bots []string
	for i := 0; i < milestones[0]; i++ {
		bots = append(bots, "sybilbot"+strconv.Itoa(i))
	}
	starAnon(cur, "sybil-app", bots...)

	l := bySlugOrPanic("sybil-app")
	contains(t, listingJSON("sybil-app"), `"stars":10`)
	uassert.Equal(t, 0, l.stars7d(blockTime()))
	uassert.False(t, strings.Contains(jsonSlugs(window(&apps.stars, 0, 100)), "sybil-app"), "not in Top")
	uassert.False(t, strings.Contains(jsonSlugs(trending(blockTime(), 10).items), "sybil-app"), "not trending")
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
	uassert.Equal(t, 0, bySlugOrPanic("mature-one").rankedCount(), "a new name's star does not rank")

	skip(nameMaturity)
	testing.SetRealm(testing.NewUserRealm(addr))
	Star(cross(cur), "mature-two")
	uassert.Equal(t, 1, bySlugOrPanic("mature-two").rankedCount())
}

func TestStars7dAndTrending(cur realm, t *testing.T) {
	setEpoch(600)
	const id = "weekly-app"
	registerFrom(cur, "gno.land/r/weekly/app", "weekly-app", "Weekly")
	l := bySlugOrPanic(id)

	starFrom(cur, id, "weeklya", "weeklyb", "weeklyc")
	skip(epochLength)
	starFrom(cur, id, "weeklyd")
	uassert.Equal(t, 4, l.stars7d(blockTime()))
	contains(t, Render("api/v1/home"), `{"title":"Trending this week","slugs":["`+id+`"`)

	skip(6 * epochLength) // the first epoch leaves the 7-epoch window
	uassert.Equal(t, 1, l.stars7d(blockTime()))
	uassert.Equal(t, 4, l.starCount())
	contains(t, Render("api/v1/home"), `{"title":"Trending this week","slugs":[]`)
}

// TestTrendingIgnoresStaleScores checks that apps that trended long ago,
// more than trendingScan of them, never hide an app trending now.
func TestTrendingIgnoresStaleScores(cur realm, t *testing.T) {
	setEpoch(800) // heights only grow until TestTrendingIsExact
	for i := 0; i <= trendingScan; i++ {
		ns := "stale" + strconv.Itoa(i)
		registerFrom(cur, "gno.land/r/"+ns+"/app", ns, "Stale "+strconv.Itoa(i))
		starFrom(cur, ns, "stalefana", "stalefanb", "stalefanc")
		if (i+1)%(maxStarsPerEpoch/3) == 0 {
			skip(epochLength) // the fans' star budget
		}
	}
	skip(30 * epochLength)
	registerFrom(cur, "gno.land/r/fresh/app", "fresh-app", "Fresh App")
	starFrom(cur, "fresh-app", "freshfana", "freshfanb")
	top := trending(blockTime(), 10).items
	uassert.Equal(t, `["fresh-app"]`, jsonSlugs(top))
}

// TestTrendingIsExact checks that an app with more stars this week trends
// ahead of apps starred more recently.
func TestTrendingIsExact(cur realm, t *testing.T) {
	setEpoch(960) // before TestBuilders, which needs the newest namespaces
	registerFrom(cur, "gno.land/r/hot/app", "hot-app", "Hot App")
	starFrom(cur, "hot-app", "hota", "hotb", "hotc", "hotd", "hote")
	skip(epochLength)
	const n = 10
	for i := 0; i < n+2; i++ {
		ns := "warm" + strconv.Itoa(i)
		registerFrom(cur, "gno.land/r/"+ns+"/app", ns, "Warm "+strconv.Itoa(i))
		starFrom(cur, ns, ns+"a", ns+"b")
	}
	top := trending(blockTime(), n).items
	uassert.Equal(t, n, len(top))
	uassert.Equal(t, "hot-app", top[0].Slug)
}

func TestPulse(cur realm, t *testing.T) {
	setEpoch(700) // earlier events are out of every window
	registerFrom(cur, "gno.land/r/pulse/one", "pulse-one", "Pulse One")
	registerFrom(cur, "gno.land/r/pulse/two", "pulse-two", "Pulse Two")
	starFrom(cur, "pulse-one", "pulsea", "pulseb", "pulsec")
	contains(t, Render("api/v1/home"), `"pulse":{"listed_7d":2,"stars_7d":3,"listed_30d":2,"stars_30d":3}`)

	skip(updateCooldown) // the first epoch leaves the 7d window
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
	contains(t, home, `"activity":[{"kind":"milestone","slug":"`+id+`","stars":10,"time":`+strconv.FormatInt(blockTime(), 10)+`}`)
	_, refs, _ := strings.Cut(home, `"listings":`)
	contains(t, refs, `"slug":"`+id+`"`)
}

func TestBuilders(cur realm, t *testing.T) {
	setEpoch(1000) // after every other listing
	registerFrom(cur, "gno.land/r/builderns/one", "builderns-one", "Builder One")
	skip(1)
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
	const now = (origin + 10000) * epochLength // every app is old by then
	old := namedApps.Size()
	first := rotation(now)
	uassert.True(t, len(first) > 0 && len(first) <= rediscoverSize)
	uassert.Equal(t, jsonSlugs(first), jsonSlugs(rotation(now+1)), "same epoch, same window")
	uassert.NotEqual(t, jsonSlugs(first), jsonSlugs(rotation(now+epochLength)), "next epoch, next window")

	// Every app of a named namespace but the core ones is in the first half
	// of a window within ceil(old/rediscoverStep) epochs; address
	// namespaces and core apps never are.
	shown := map[string]bool{}
	for e := int64(0); e < int64((old+rediscoverStep-1)/rediscoverStep); e++ {
		w := rotation(now + e*epochLength)
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
	writeListing(&b, bySlugOrPanic(slug), blockTime())
	return b.String()
}

// ev is the event of kind for the listing with this slug, at the current time.
func ev(kind, slug string, stars int) event {
	return event{kind, bySlugOrPanic(slug), stars, blockTime()}
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
	for _, s := range homeShelves(blockTime()) {
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
	for _, s := range homeShelves(blockTime()) {
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
	now := blockTime()
	uassert.ErrorContains(t, setPick("nope", 7, now), "unknown listing")
	uassert.ErrorContains(t, setPick("avl", 7, now), "only an app")
	uassert.ErrorContains(t, setPick("blog", 0, now), "1 to 30 days")
	uassert.ErrorContains(t, setPick("blog", maxPickDays+1, now), "1 to 30 days")
	uassert.AbortsContains(t, cur, "1 to 30 days", func() { ProposeSpotlight(cross(cur), "blog", 31) })

	uassert.NoError(t, setPick("boards", 7, now))
	uassert.True(t, picked(now) == bySlugOrPanic("boards"))
	contains(t, Render("api/v1/home"), `"pick":"boards"`)
	contains(t, Render(""), "## Picked by GovDAO")
	uassert.True(t, picked(now+7*day) == nil, "the pick ends")

	uassert.NoError(t, hide("boards", "other"))
	uassert.True(t, picked(now) == nil, "a hidden app is no longer picked")
	uassert.NoError(t, unhide("gno.land/r/gnoland/boards2/v0"))
	uassert.True(t, picked(now) == nil, "nor once unhidden: the hide ended the pick")
	ProposeSpotlight(cross(cur), "boards", 7)
}

// TestPickEndsWithTheShownApp checks that a pick ends for good when its app
// stops being a shown app, and that it is bound to the title it was picked
// under.
func TestPickEndsWithTheShownApp(cur realm, t *testing.T) {
	owner := testutils.TestAddress("heroowner")
	path := "gno.land/r/" + owner.String() + "/hero"
	registerFrom(cur, path, "hero-app", "Hero App")
	now := blockTime()

	// A kind change app -> service -> app.
	uassert.NoError(t, setPick("hero-app", 7, now))
	testing.SetRealm(testing.NewUserRealm(owner))
	Claim(cross(cur), path, "hero-app", "Hero App", "A tagline", "utilities", 4, kindService)
	uassert.True(t, picked(now) == nil, "a service is not picked")
	Claim(cross(cur), path, "hero-app", "Hero App", "A tagline", "utilities", 4, kindApp)
	uassert.True(t, picked(now) == nil, "nor once an app again")

	// A rename, even undone.
	uassert.NoError(t, setPick("hero-app", 7, now))
	uassert.True(t, picked(now) == bySlugOrPanic("hero-app"))
	registerFrom(cur, path, "hero-app", "Hero App Renamed")
	uassert.True(t, picked(now) == nil, "a renamed app is not the one picked")
	registerFrom(cur, path, "hero-app", "Hero App")
	uassert.True(t, picked(now) == nil, "nor once renamed back")

	// A proposal is bound to the app as it was shown when proposed.
	l := bySlugOrPanic("hero-app")
	shown := pickedAs(l, now)
	testing.SetRealm(testing.NewCodeRealm(path))
	Register(cross(cur), "", "Hero App", "Official GovDAO airdrop, claim yours now", "utilities", 4)
	uassert.ErrorContains(t, pickBound("hero-app", path, shown, 7, now), "changed since the proposal")
	uassert.NoError(t, pickBound("hero-app", path, pickedAs(l, now), 7, now))
	uassert.True(t, picked(now) == l)

	// Once picked, a new tagline or new art ends the pick.
	registerFrom(cur, path, "hero-app", "Hero App")
	uassert.True(t, picked(now) == nil, "a new tagline is not what was voted on")
	uassert.NoError(t, setPick("hero-app", 7, blockTime()))
	testing.SetRealm(testing.NewCodeRealm(path))
	SubmitRich(cross(cur), testIcon, "")
	uassert.True(t, picked(blockTime()) == l, "a change still pending is not shown yet")
	skip(richDelay)
	uassert.True(t, picked(blockTime()) == nil, "new art is not what was voted on")
}

// TestNewSkipsAddressNamespaces checks that an app of a g1… namespace, even
// the newest, is not on the front's New but is on its list, :latest.
func TestNewSkipsAddressNamespaces(cur realm, t *testing.T) {
	setEpoch(2000) // after every other listing
	path := "gno.land/r/" + testutils.TestAddress("newanon").String() + "/app"
	registerFrom(cur, path, "anon-new", "Anon New")
	skip(1)
	registerFrom(cur, "gno.land/r/newnamed/app", "named-new", "Named New")

	front := homeShelves(blockTime())[0]
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
	setEpoch(3000) // after every other listing
	registerFrom(cur, "gno.land/r/gnoland/extra", "gnoland-extra", "Gnoland Extra")
	starFrom(cur, "gnoland-extra", "seedfana", "seedfanb", "seedfanc")
	skip(1)
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
	uassert.ErrorContains(t, pickBound("bound-app", victim, pickedAs(bySlugOrPanic("bound-app"), blockTime()), 7, blockTime()), "another listing")
	uassert.False(t, bySlugOrPanic("bound-app").hidden)
	uassert.True(t, picked(blockTime()) == nil || picked(blockTime()).Path != squatter)
	uassert.NoError(t, pickBound("bound-app", squatter, pickedAs(bySlugOrPanic("bound-app"), blockTime()), 1, blockTime()))
	uassert.NoError(t, hideBound("bound-app", squatter, "other"))
}

// TestPickNeedsConfirmed checks that a claimed app no one confirmed cannot
// be picked: picked() would not show it.
func TestPickNeedsConfirmed(cur realm, t *testing.T) {
	owner := testutils.TestAddress("pickowner")
	testing.SetRealm(testing.NewUserRealm(owner))
	Claim(cross(cur), "gno.land/r/"+owner.String()+"/app", "unconfirmed-pick", "Unconfirmed Pick", "A tagline", "nft-art", 1, kindApp)
	uassert.ErrorContains(t, setPick("unconfirmed-pick", 7, blockTime()), "only a confirmed app")
	uassert.AbortsContains(t, cur, "only a confirmed app", func() { ProposeSpotlight(cross(cur), "unconfirmed-pick", 7) })
}

// TestBlock checks the embeddable lists: the front's rules, n clamped, the
// titles sanitized, and nothing for an unknown list.
func TestBlock(cur realm, t *testing.T) {
	setEpoch(4000) // after every other listing
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

// A free g1… namespace earns no front-page slot: an update keeps its app on
// :updated but off Recently updated, and out of the feed.
func TestFrontLeavesAddressNamespacesOut(cur realm, t *testing.T) {
	setEpoch(5000) // after every other listing
	path := "gno.land/r/" + testutils.TestAddress("freeanon").String() + "/app"
	registerFrom(cur, path, "anon-app", "Anon App")
	skip(updateCooldown)
	testing.SetRealm(testing.NewCodeRealm(path))
	Register(cross(cur), "", "Anon App", "Changed tagline", "utilities", 4)

	l := bySlugOrPanic("anon-app")
	uassert.True(t, apps.updated.Has(descKey(l.Updated, l.Slug)), ":updated lists every app")
	for _, e := range activity.newestFirst() {
		uassert.True(t, e.Listing != l, "not in the feed")
	}
	uassert.False(t, strings.Contains(Render(""), "Anon App"), "on no front-page shelf")
	uassert.False(t, strings.Contains(Render("api/v1/home"), `"slug":"anon-app"`))
}

// Listings the feed never serves cannot push an app's event out of it.
func TestFeedKeepsAppEvents(cur realm, t *testing.T) {
	registerFrom(cur, "gno.land/r/feedapp/app", "feed-app", "Feed App")
	for i := 0; i < activitySize; i++ {
		n := strconv.Itoa(i)
		owner := testutils.TestAddress("svcflood" + n)
		path := "gno.land/r/" + owner.String() + "/svc"
		testing.SetRealm(testing.NewUserRealm(owner))
		Claim(cross(cur), path, "flood-svc-"+n, "Flood Service "+n, "A tagline", "infrastructure", 1, kindService)
		registerIn(cur, path, "ignored", "Flood Service "+n, "infrastructure")
	}
	uassert.Equal(t, "feed-app", activity.newestFirst()[0].Listing.Slug)
}

// Every slot of the trending window is read: a crowd of tied two-star apps
// filed one day cannot hide the week's top app, filed another.
func TestTrendingReadsEverySlot(cur realm, t *testing.T) {
	const e = 5010 // (origin+e)%trendEpochs == 6: the last slot read
	setEpoch(e)
	registerFrom(cur, "gno.land/r/hotslot/app", "hotslot-app", "Hot Slot")
	starFrom(cur, "hotslot-app", "hsa", "hsb", "hsc", "hsd", "hse")
	setEpoch(e + 1) // slot 0, read first
	for i := 0; i < trendingScan; i++ {
		ns, fan := "slotfill"+strconv.Itoa(i), "fill"+strconv.Itoa(i/10)
		registerFrom(cur, "gno.land/r/"+ns+"/app", ns, "Slot Fill "+strconv.Itoa(i))
		starFrom(cur, ns, fan+"a", fan+"b")
	}
	top := trending(blockTime(), 10)
	uassert.True(t, len(top.items) > 0 && top.items[0].Slug == "hotslot-app", jsonSlugs(top.items))
}

// A core listing is frozen in code: its namespace's owner cannot rewrite it.
func TestClaimRefusesCoreListing(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(nameUser(cur, "demo")))
	uassert.AbortsContains(t, cur, "core listing cannot be claimed", func() {
		Claim(cross(cur), "gno.land/r/demo/defi/grc20factory", "token-factory", "Free GNOT Airdrop", "A tagline", "defi", 1, kindApp)
	})
	uassert.Equal(t, "Token Factory", bySlugOrPanic("token-factory").Title)
}

// An app's own Register takes its slug and title from an unconfirmed claim
// holding them: the app's init runs once, the claim proves nothing.
func TestRegisterEvictsSquattingClaim(cur realm, t *testing.T) {
	squatter := testutils.TestAddress("squatter")
	squat := "gno.land/r/" + squatter.String() + "/squat"
	testing.SetRealm(testing.NewUserRealm(squatter))
	Claim(cross(cur), squat, "real-app", "Real App", "A tagline", "utilities", 1, kindApp)

	registerFrom(cur, "gno.land/r/realns/app", "real-app", "Real App")
	uassert.Equal(t, "gno.land/r/realns/app", bySlugOrPanic("real-app").Path)
	uassert.False(t, byPath.Has(squat), "the claim is dropped")

	// A confirmed listing still holds its title.
	registerFrom(cur, "gno.land/r/realcopy/app", "real-copy", "Real App")
	uassert.False(t, listings.Has("real-copy"))

	// A refused Register takes nothing: a claim survives a Register that
	// names its slug and a confirmed listing's title...
	claimed := "gno.land/r/" + squatter.String() + "/other"
	testing.SetRealm(testing.NewUserRealm(squatter))
	Claim(cross(cur), claimed, "kept-claim", "Kept Claim", "A tagline", "utilities", 1, kindApp)
	registerFrom(cur, "gno.land/r/thief/app", "kept-claim", "Boards")
	uassert.True(t, byPath.Has(claimed), "the claim survives a refused Register")
	// ...and a hidden listing's Register, which takes no title.
	registerFrom(cur, "gno.land/r/hiddenthief/app", "hidden-thief", "Hidden Thief")
	uassert.NoError(t, hide("hidden-thief", "phishing"))
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/hiddenthief/app"))
	Register(cross(cur), "", "Kept Claim", "A tagline", "utilities", 4)
	uassert.True(t, byPath.Has(claimed), "a hidden listing evicts nothing")
}

// A claim confirmed by stars is a claim again once its backers leave; a
// listing its own realm listed stays confirmed.
func TestClaimLosesConfirmationWithItsStars(cur realm, t *testing.T) {
	owner := testutils.TestAddress("backed")
	testing.SetRealm(testing.NewUserRealm(owner))
	Claim(cross(cur), "gno.land/p/"+owner.String()+"/lib", "backed-lib", "Backed Lib", "A tagline", "dev-tools", 1, kindPackage)
	starFrom(cur, "backed-lib", "backfana", "backfanb", "backfanc")
	l := bySlugOrPanic("backed-lib")
	uassert.True(t, l.Confirmed)
	testing.SetRealm(testing.NewUserRealm(nameUser(cur, "backfana")))
	Unstar(cross(cur), "backed-lib")
	uassert.False(t, l.Confirmed)
	uassert.False(t, toolbox.created.Has(descKey(l.Created, l.Slug)), "off the shelves")

	registerFrom(cur, "gno.land/r/selflisted/app", "self-listed", "Self Listed")
	starFrom(cur, "self-listed", "selffana", "selffanb", "selffanc")
	testing.SetRealm(testing.NewUserRealm(nameUser(cur, "selffana")))
	Unstar(cross(cur), "self-listed")
	uassert.True(t, bySlugOrPanic("self-listed").Confirmed)
}

// A star given before the name counted ranks once starred again.
func TestStarAgainOnceTheNameCounts(cur realm, t *testing.T) {
	registerFrom(cur, "gno.land/r/starlater/app", "star-later", "Star Later")
	fan := nameUser(cur, "laterfan")
	testing.SetRealm(testing.NewUserRealm(fan))
	Star(cross(cur), "star-later")
	l := bySlugOrPanic("star-later")
	uassert.Equal(t, 0, l.rankedCount(), "a name's first star only starts its clock")

	skip(nameMaturity)
	testing.SetRealm(testing.NewUserRealm(fan))
	Star(cross(cur), "star-later")
	uassert.Equal(t, 1, l.rankedCount())
	uassert.Equal(t, 1, l.starCount())
	uassert.AbortsContains(t, cur, "already starred", func() { Star(cross(cur), "star-later") })
}

// A pause stops every listing update, the icon and cover included.
func TestPauseStopsSubmitRich(cur realm, t *testing.T) {
	registerFrom(cur, "gno.land/r/pauserich/app", "pause-rich", "Pause Rich")
	paused = true
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/pauserich/app"))
	uassert.AbortsContains(t, cur, "paused", func() { SubmitRich(cross(cur), testIcon, "") })
	paused = false
}

// The core namespaces hold the seed: the cap leaves them room.
func TestCoreNamespacesHaveNoCap(cur realm, t *testing.T) {
	for i := 0; i < maxPerNamespace; i++ {
		n := strconv.Itoa(i)
		registerFrom(cur, "gno.land/r/gnoland/extra"+n, "gnoland-extra-"+n, "Gnoland Extra "+n)
		uassert.True(t, listings.Has("gnoland-extra-"+n), n)
	}
}

// A successor points readers, embedding realms and gnoweb to the new store.
func TestSuccessor(cur realm, t *testing.T) {
	uassert.AbortsContains(t, cur, "a successor is a realm path", func() { ProposeSuccessor(cross(cur), "gno.land/p/gnoland/store/v1") })
	uassert.AbortsContains(t, cur, "a successor is a realm path", func() { ProposeSuccessor(cross(cur), "/r/gnoland/store/v1") })
	ProposeSuccessor(cross(cur), "gno.land/r/gnoland/store/v1")

	successor = "gno.land/r/gnoland/store/v1"
	defer func() { successor = "" }()
	moved := "Explore has moved: [/r/gnoland/store/v1](/r/gnoland/store/v1)"
	contains(t, Render(""), moved)
	contains(t, Render("top"), moved)
	contains(t, Block("latest", 1), moved)
	uassert.Equal(t, moved, Badge("gno.land/r/gnoland/blog"))
	contains(t, Render("api/v1/home"), `"moved_to":"gno.land/r/gnoland/store/v1"`)
}

// Only the seed's listings are core: gnoweb's Core badge reads this flag,
// not its trust list.
func TestCoreFlag(t *testing.T) {
	contains(t, listingJSON("blog"), `"core":true`)
	contains(t, listingJSON("starred-app"), `"core":false`)
}

// A shelf cut from a computed list keeps one count per app.
func TestTrendingShelfKeepsItsCounts(cur realm, t *testing.T) {
	setEpoch(6000) // after every other trending app
	for i := 0; i <= gridPool; i++ {
		ns := "shelfcut" + strconv.Itoa(i)
		registerFrom(cur, "gno.land/r/"+ns+"/app", ns, "Shelf Cut "+strconv.Itoa(i))
		starFrom(cur, ns, "shelfcuta", "shelfcutb")
	}
	s := listShelf(mustList("trending", blockTime()))
	uassert.Equal(t, gridPool, len(s.Items))
	uassert.Equal(t, len(s.Items), len(s.Counts))
}

// Hiding a namespace's oldest listing moves its builder's First, in the
// ranking ordered by it too.
func TestBuilderFirstFollowsTheOldest(cur realm, t *testing.T) {
	registerFrom(cur, "gno.land/r/firstns/one", "firstns-one", "First One")
	skip(1)
	registerFrom(cur, "gno.land/r/firstns/two", "firstns-two", "First Two")
	one, two := bySlugOrPanic("firstns-one"), bySlugOrPanic("firstns-two")
	b := builders.Get("firstns").(*builder)
	uassert.Equal(t, one.Created, b.First)

	uassert.NoError(t, hide("firstns-one", "other"))
	uassert.Equal(t, two.Created, b.First)
	uassert.True(t, buildersByFirst.Has(descKey(two.Created, "firstns")))
	uassert.False(t, buildersByFirst.Has(descKey(one.Created, "firstns")))
}

// Unstarring a star older than the trending window leaves the window's
// count alone.
func TestUnstarAnOldStar(cur realm, t *testing.T) {
	registerFrom(cur, "gno.land/r/oldstar/app", "old-star", "Old Star")
	l := bySlugOrPanic("old-star")
	starFrom(cur, "old-star", "oldstara")
	skip((trendEpochs + 1) * epochLength)
	starFrom(cur, "old-star", "oldstarb")
	uassert.Equal(t, 1, l.stars7d(blockTime()))

	testing.SetRealm(testing.NewUserRealm(nameUser(cur, "oldstara")))
	Unstar(cross(cur), "old-star")
	uassert.Equal(t, 1, l.stars7d(blockTime()))
	uassert.Equal(t, 1, l.rankedCount())
}

// A claim confirmed again after its backers left is not announced again:
// unstarring and starring in a loop buys no feed slot.
func TestReconfirmedClaimIsAnnouncedOnce(cur realm, t *testing.T) {
	owner := nameUser(cur, "loopowner")
	testing.SetRealm(testing.NewUserRealm(owner))
	Claim(cross(cur), "gno.land/r/loopowner/app", "loop-app", "Loop App", "A tagline", "utilities", 1, kindApp)
	starFrom(cur, "loop-app", "loopfana", "loopfanb", "loopfanc")
	l := bySlugOrPanic("loop-app")
	uassert.True(t, l.Confirmed)
	listed := pulseListed.sum(blockTime()/epochLength, 7)

	fan := nameUser(cur, "loopfana")
	for i := 0; i < 3; i++ {
		testing.SetRealm(testing.NewUserRealm(fan))
		Unstar(cross(cur), "loop-app")
		testing.SetRealm(testing.NewUserRealm(fan))
		Star(cross(cur), "loop-app")
	}
	uassert.True(t, l.Confirmed)
	uassert.Equal(t, listed, pulseListed.sum(blockTime()/epochLength, 7))
	n := 0
	for _, e := range activity.newestFirst() {
		if e.Listing == l && e.Kind == "listed" {
			n++
		}
	}
	uassert.Equal(t, 1, n)
}

// A claim confirmed while hidden is announced once GovDAO unhides it.
func TestUnhiddenClaimIsAnnounced(cur realm, t *testing.T) {
	owner := nameUser(cur, "hiddenowner")
	path := "gno.land/r/hiddenowner/app"
	testing.SetRealm(testing.NewUserRealm(owner))
	Claim(cross(cur), path, "hidden-claim", "Hidden Claim", "A tagline", "utilities", 1, kindApp)
	uassert.NoError(t, hide("hidden-claim", "other"))
	registerIn(cur, path, "ignored", "Hidden Claim", "utilities") // confirmed while hidden
	uassert.NoError(t, unhide(path))

	l := lookup(&byPath, path)
	uassert.True(t, l.Confirmed && l.announced)
	uassert.True(t, activity.newestFirst()[0].Listing == l)
}
