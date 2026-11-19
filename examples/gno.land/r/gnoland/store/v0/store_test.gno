package store

import (
	"chain/runtime"
	"strconv"
	"strings"
	"testing"
	"time"
	"unicode"
	"unicode/utf8"

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
		testing.SetRealm(testing.NewUserRealm(rankedFan(cur, u)))
		Star(cross(cur), slug)
	}
}

// rankedFan returns the address of user u, named and known to the store for
// nameMaturity: its stars rank.
func rankedFan(cur realm, u string) address {
	addr := nameUser(cur, u)
	if !voters.Has(addr.String()) {
		voters.Set(addr.String(), voter{NameSeen: -nameMaturity})
	}
	return addr
}

// enableNames turns r/sys/names on, as its admin: names then own namespaces.
func enableNames(cur realm) {
	if !names.IsEnabled() {
		testing.SetRealm(testing.NewUserRealm("g1skl80cuz8zq3lul9pgz5pc35l2pfzgxgfpsqkx"))
		names.Enable(cross(cur))
	}
}

// hide hides the visible listing slug for reason, as a proposal built by
// ProposeHide does once executed.
func hide(slug, reason string) error {
	if err := checkHide(slug, reason); err != nil {
		return err
	}
	return hidePath(lookup(&listings, slug).Path, reason)
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
	contains(t, listingJSON("acme-swap"), `{"slug":"acme-swap","kind":"app","path":"gno.land/r/acme/swap","title":"Acme Swap","tagline":"A tagline","category":"utilities","palette":4,"stars":0,"ranked":0,"stars_7d":0,"fifth_ranked":0,"created":`+strconv.FormatInt(l.Created, 10)+`,"updated":`+strconv.FormatInt(l.Updated, 10)+`,"pending":false,"core":false,"confirmed":true,"icon":"","cover":""}`)
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
	shelf := updatedShelf()

	for i := 0; i < 5; i++ {
		skip(1)
		registerSpam(cur, "Version 1") // nothing changes
	}
	registerSpam(cur, "Version 2") // a real change, too soon
	uassert.Equal(t, "Version 2", l.Tagline, "the change is applied")
	uassert.Equal(t, l.Created, l.Updated, "but not counted as an update")
	uassert.Equal(t, shelf, updatedShelf())

	testing.SetRealm(testing.NewCodeRealm(spamPath))
	SubmitRich(cross(cur), testIcon, "")
	uassert.Equal(t, l.Created, l.Updated, "SubmitRich obeys the same cooldown")

	skip(updateCooldown)
	registerSpam(cur, "Version 2") // unchanged: still no bump
	uassert.Equal(t, l.Created, l.Updated)
	registerSpam(cur, "Version 3")
	bumped := blockTime()
	uassert.Equal(t, bumped, l.Updated)
	skip(1)
	registerSpam(cur, "Version 4")
	uassert.Equal(t, "Version 4", l.Tagline)
	uassert.Equal(t, bumped, l.Updated, "one update per cooldown")

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
		{"moved mark", "my--slug", "Title", "Tagline", "defi", 0, "bad_slug"},
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
	// Of the typographic quotes, a title holds the four titleKey folds: the
	// low and reversed ones would key apart from the mark they draw as.
	for r := '‘'; r <= '‟'; r++ {
		uassert.Equal(t, foldRune(r) != r, validTitle("Acme"+string(r)+" Swap"), string(r))
	}
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
		{"Gnoland", "Gnoأand"},                               // alef with hamza
		{"ПОРТ", "ΠΟΡΤ"}, {"ГОРА", "ΓΟΡΑ"}, {"ПОРТ", "ΠОРТ"}, // Greek and Cyrillic, alike or mixed
		{"カメラ", "力メラ"}, {"ヘルプ", "へルプ"}, // Han and hiragana drawn as katakana
		{"Bob's Swap", "Bob’s Swap"},                                                             // a typographic apostrophe
		{"Bob's Swap", "Bob`s Swap"}, {"Bob‘s Swap", "Bob‛s Swap"}, {"Acme, Swap", "Acme‚ Swap"}, // a grave accent, the low and reversed quotes
		{"Open", "Ореп"}, {"Open", "Ορεη"}, {"Book", "Ьоок"}, {"Tor", "Тог"}, {"Tor", "Тоґ"}, {"Tor", "Тоѓ"}, // Cyrillic pe, soft sign and ghe, Greek eta
		{"HOT", "ΗΟΤ"}, // whose capital stays an H
		// Found by an adversarial pass against Unicode's confusables, each
		// drawn as its original, or within a pixel at text size.
		{"\u91D1\u878D Swap", "\uF90A\u878D Swap"},                                 // cjk-compat
		{"\u4EA4\u6613\u6240", "\u4EA4\uF9E0\u6240"},                               // cjk-compat
		{"\u092C\u093E\u091C\u093C\u093E\u0930", "\u092C\u093E\u095B\u093E\u0930"}, // nukta
		{"\u0E41\u0E21\u0E27", "\u0E40\u0E40\u0E21\u0E27"},                         // thai-sara-ae
		{"\u0643\u062A\u0627\u0628", "\u06A9\u062A\u0627\u0628"},                   // keheh
		{"\u0628\u064A\u062A", "\u0628\u06CC\u062A"},                               // farsi-yeh
		{"\u0639\u0644\u064A", "\u0639\u0644\u0649"},                               // alef-maksura
		{"\u0420\u042B\u041D\u041E\u041A", "\u0420\u042C\u0406\u041D\u041E\u041A"}, // yeru
		{"The \"Real\" Swap", "The ''Real'' Swap"},                                 // double-quote
		{"ACME 3", "\u0410\u0421\u041C\u0415 \u0417"},                              // ze
		{"\u0422\u041E\u041F 3", "\u0422\u041E\u041F \u0417"},                      // ze
		{"\u30D2\u30C3\u30C8", "\u5315\u30C3\u30C8"},                               // han-hi
		{"\u30A4\u30AB", "\u4EBB\u30AB"},                                           // han-i
		{"\u30E0\u30E9", "\u53B6\u30E9"},                                           // han-mu
		{"Bob's Swap", "Bob\u05D9s Swap"},                                          // yod
		{"Rock 'n Roll", "Rock \u0149 Roll"},                                       // n-apostrophe
		{"\u043A\u043E\u043D\u044C", "\u043A\u043E\u045A"},                         // nje
		{"\u041E\u0431\u044C\u0435\u043C", "\u041E\u0431\u044A\u0435\u043C"},       // hard-sign
		{"Swap", "\u093Dwap"},                                                      // avagraha
		{"TOKEN", "\u4E05OKEN"},                                                    // han-t
		{"TOKEN", "\u4E01OKEN"},                                                    // han-ding
		{"ZONE", "\u4E59ONE"},                                                      // han-yi
		{"Gno+", "Gno\u5341"},                                                      // han-ten
		{"\u53E3\u30B3\u30DF", "\u56D7\u30B3\u30DF"},                               // han-mouth
		{"Gnoland", "Gno\u0672and"},                                                // alef-wavy
		{"Gnoland", "Gno\u0675and"},                                                // alef-high
		{"Why?", "Why\u097D"},                                                      // glottal
		{"Block Explorer", "\u042B\u043E\u0441\u043A \u0415\u0445\u0440\u0406\u043E\u0433\u0435\u0433"}, // yeru-bl
		{"Area 6", "\u0410\u0433\u0435\u0430 \u0431"},                                                   // be
		{"YES", "\u4E2BES"},                                // han-y
		{"AC/DC", "AC\u30CEDC"},                            // katakana-no
		{"AC/DC", "AC\u4E3FDC"},                            // han-pie
		{"Gnoland", "Gn\u06BEland"},                        // heh-doachashmee
		{"Wallet", "Wa\u05F0et"},                           // double-vav
		{"The \"Real\" Swap", "The \u05F2Real\u05F2 Swap"}, // double-yod
		{"\u30AB\u30E1\u30E9", "\uFF76\uFF92\uFF97"},       // halfwidth
		{"\u4E3D\u6C5F", "\U0002F800\u6C5F"},               // cjk-compat-supp
		{"Gnoland", "Gno\U0001EE00and"},                    // arabic-math
	} {
		uassert.True(t, !validTitle(pair[1]) || titleKey(pair[1]) == titleKey(pair[0]), pair[1])
	}
	// A CJK compatibility ideograph is a second code point of an ideograph,
	// drawn as it: a title holds none.
	for _, block := range [][2]rune{{0xF900, 0xFAFF}, {0x2F800, 0x2FA1F}} {
		for r := block[0]; r <= block[1]; r++ {
			uassert.Equal(t, isUnifiedIdeograph(r), validTitle("\u91D1"+string(r)), strconv.Itoa(int(r)))
		}
	}
	// Twelve of the block are letters of their own, as in names: 山﨑.
	uassert.True(t, validTitle("\u5C71\uFA11"))
	uassert.Equal(t, titleKey("\u5C71\u5D0E"), titleKey("\u5C71\uFA11"), "and keys as 崎, which it draws like")
	uassert.Equal(t, titleKey("\u03B7"), titleKey("\u03AE"), "eta with or without tonos")
}

// The prolonged sound mark, which most Japanese loanwords hold, is a letter
// of the script Common: a title holds it, and it keys as the dash it draws
// as, like the Han one.
func TestJapaneseLongVowel(t *testing.T) {
	for _, title := range []string{"\u30B2\u30FC\u30E0", "\u30C8\u30FC\u30AF\u30F3", "\u30E6\u30FC\u30B6\u30FC", "\u3089\u30FC\u3081\u3093", "\u30B2\u30FC\u30E0 Swap"} {
		uassert.True(t, validTitle(title), title)
	}
	uassert.Equal(t, titleKey("\u30B2\u30FC\u30E0"), titleKey("\u30B2\u4E00\u30E0"))
	uassert.False(t, validTitle("\uFF79\uFF70\uFF91"), "halfwidth katakana")
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
	enableNames(cur)

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
		_, shelf, _ := strings.Cut(Render("api/v1/build"), `{"key":"new"`)
		shelf, _, _ = strings.Cut(shelf, "]")
		return shelf
	}
	uassert.False(t, l.Confirmed)
	contains(t, listingJSON(slug), `"confirmed":false`)
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
// reaches the shelves once its own Register confirms it.
func TestClaimedAppConfirmedByRegister(cur realm, t *testing.T) {
	owner := testutils.TestAddress("appowner")
	path := "gno.land/r/" + owner.String() + "/app"
	testing.SetRealm(testing.NewUserRealm(owner))
	Claim(cross(cur), path, "owned-app", "Owned App", "A tagline", "nft-art", 1, kindApp)
	l := bySlugOrPanic("owned-app")
	uassert.False(t, l.Confirmed)
	uassert.False(t, byTitle.Has(titleKey("Owned App")), "a claim holds no title")
	contains(t, Render("api/v1/category/nft-art/1"), `"slug":"owned-app"`)

	registerIn(cur, path, "ignored", "Owned App", "nft-art")
	uassert.True(t, l.Confirmed)
	uassert.True(t, apps.created.Has(descKey(l.Created, l.Slug)), "on the shelves")
	uassert.Equal(t, "owned-app", titleHolder(l), "confirmed, it holds its title")
}

// TestUnconfirmedAnnouncesNothing checks that a claim stays out of the
// pulse, even when updated.
func TestUnconfirmedAnnouncesNothing(cur realm, t *testing.T) {
	owner := testutils.TestAddress("ghostowner")
	path := "gno.land/r/" + owner.String() + "/ghost"
	epoch := blockTime() / epochLength
	listed := pulseListed.sum(epoch, 30)
	testing.SetRealm(testing.NewUserRealm(owner))
	Claim(cross(cur), path, "ghost-app", "Ghost App", "A tagline", "nft-art", 1, kindApp)
	skip(updateCooldown)
	Claim(cross(cur), path, "ghost-app", "Ghost App", "Another tagline", "nft-art", 1, kindApp)
	uassert.Equal(t, "Another tagline", bySlugOrPanic("ghost-app").Tagline)
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

	// A core listing keeps its slug and its title while hidden.
	registerFrom(cur, "gno.land/r/coinsquat/app", "coins", "Coins Squat")
	registerFrom(cur, "gno.land/r/coinsquat/app", "coin-squat", "Coin Balances")
	uassert.False(t, byPath.Has("gno.land/r/coinsquat/app"), "neither the slug nor the title of a hidden core listing is free")

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

// TestHideReleasesSlug checks that hiding a squatter frees its slug and its
// title for the real app, for good. An unhide gives neither back: it waits
// for another title, which the hidden realm can take, and brings the listing
// back under a slug of its own, since no one can register one to keep it
// hidden.
func TestHideReleasesSlug(cur realm, t *testing.T) {
	const squat = "gno.land/r/squatter/app"
	registerFrom(cur, squat, "brand-swap", "Brand Swap")
	uassert.NoError(t, hide("brand-swap", "impersonation"))
	uassert.ErrorContains(t, hide("brand-swap", "other"), "unknown listing")

	registerFrom(cur, "gno.land/r/brand/swap", "brand-swap", "Brand Swap")
	real := bySlugOrPanic("brand-swap")
	uassert.Equal(t, "gno.land/r/brand/swap", real.Path)

	l := lookup(&byPath, squat)
	movedBefore, visibleBefore := moved, listings.Size()
	uassert.ErrorContains(t, unhide(squat), "store: title_taken")
	uassert.True(t, l.hidden && l.Slug == "brand-swap" && moved == movedBefore && listings.Size() == visibleBefore,
		"a refused unhide changes nothing")
	registerFrom(cur, squat, "", "Brand Fork")
	uassert.NoError(t, unhide(squat))
	uassert.True(t, strings.HasPrefix(l.Slug, "brand-swap--"), l.Slug)
	uassert.True(t, bySlugOrPanic(l.Slug) == l && bySlugOrPanic("brand-swap") == real)
	uassert.Equal(t, l.Slug, titleHolder(l))
	uassert.True(t, lookup(&apps.created, descKey(l.Created, l.Slug)) == l)

	registerFrom(cur, squat, "brand-swap", "Brand Fork") // still held
	uassert.True(t, strings.HasPrefix(l.Slug, "brand-swap--"), l.Slug)
	registerFrom(cur, squat, "brand-fork", "Brand Fork")
	uassert.Equal(t, "brand-fork", l.Slug)

	// A claim's slug may go to another claim while it is hidden.
	owner, other := testutils.TestAddress("hidslugowner"), testutils.TestAddress("hidslugother")
	path := "gno.land/p/" + owner.String() + "/lib"
	testing.SetRealm(testing.NewUserRealm(owner))
	Claim(cross(cur), path, "hid-lib", "Hid Lib", "A tagline", "dev-tools", 1, kindPackage)
	uassert.NoError(t, hide("hid-lib", "other"))
	testing.SetRealm(testing.NewUserRealm(other))
	Claim(cross(cur), "gno.land/p/"+other.String()+"/lib", "hid-lib", "Other Lib", "A tagline", "dev-tools", 1, kindPackage)
	uassert.NoError(t, unhide(path))
	uassert.True(t, strings.HasPrefix(lookup(&byPath, path).Slug, "hid-lib--"))
	uassert.Equal(t, "gno.land/p/"+other.String()+"/lib", bySlugOrPanic("hid-lib").Path)

	// A confirmed listing whose slug a claim took while it was hidden comes
	// back under it, and the claim moves, as on a Register.
	registerFrom(cur, "gno.land/r/hidback/app", "hid-back", "Hid Back")
	back := bySlugOrPanic("hid-back")
	uassert.NoError(t, hide("hid-back", "other"))
	testing.SetRealm(testing.NewUserRealm(other))
	Claim(cross(cur), "gno.land/r/"+other.String()+"/back", "hid-back", "Hid Claim", "A tagline", "utilities", 1, kindApp)
	claim := bySlugOrPanic("hid-back")
	uassert.NoError(t, unhide("gno.land/r/hidback/app"))
	uassert.True(t, bySlugOrPanic("hid-back") == back && back.Slug == "hid-back")
	uassert.True(t, strings.HasPrefix(claim.Slug, "hid-back--"), claim.Slug)
	uassert.True(t, bySlugOrPanic(claim.Slug) == claim)
}

// A claim with the stars to be confirmed, hidden while another listing
// holds its title, is not confirmed by its owner's Claim: it would then need
// that title to come back.
func TestHiddenClaimWaitsForItsUnhide(cur realm, t *testing.T) {
	owner := testutils.TestAddress("waitclaim")
	path := "gno.land/p/" + owner.String() + "/lib"
	testing.SetRealm(testing.NewUserRealm(owner))
	Claim(cross(cur), path, "wait-claim", "Wait Title", "A tagline", "dev-tools", 1, kindPackage)
	registerFrom(cur, "gno.land/r/waitholder/app", "wait-holder", "Wait Title")
	starFrom(cur, "wait-claim", "waitfana", "waitfanb", "waitfanc")
	claim := lookup(&byPath, path)
	uassert.NoError(t, hide("wait-claim", "other"))
	testing.SetRealm(testing.NewUserRealm(owner))
	Claim(cross(cur), path, "wait-claim", "Wait Title", "A tagline", "dev-tools", 1, kindPackage)
	uassert.False(t, claim.Confirmed)
	uassert.NoError(t, unhide(path))
	uassert.False(t, claim.Confirmed, "its title is still held")
	registerFrom(cur, "gno.land/r/waitholder/app", "", "Wait Holder")
	testing.SetRealm(testing.NewUserRealm(owner))
	Claim(cross(cur), path, "wait-claim", "Wait Title", "A tagline", "dev-tools", 1, kindPackage)
	uassert.True(t, claim.Confirmed)
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
	for i := 0; i < 10; i++ {
		bots = append(bots, "sybilbot"+strconv.Itoa(i))
	}
	starAnon(cur, "sybil-app", bots...)

	l := bySlugOrPanic("sybil-app")
	contains(t, listingJSON("sybil-app"), `"stars":10`)
	uassert.Equal(t, 0, l.stars7d(blockTime()))
	uassert.False(t, strings.Contains(jsonSlugs(window(&apps.stars, 0, 100)), "sybil-app"), "not in Top")
	uassert.False(t, strings.Contains(jsonSlugs(trending(blockTime(), 10).items), "sybil-app"), "not trending")
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
	contains(t, Render("api/v1/home"), `{"key":"trending","slugs":["`+id+`"`)

	skip(6 * epochLength) // the first epoch leaves the 7-epoch window
	uassert.Equal(t, 1, l.stars7d(blockTime()))
	uassert.Equal(t, 4, l.starCount())
	contains(t, Render("api/v1/home"), `{"key":"trending","slugs":[]`)
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
		`{"key":"new","slugs":["`,
		`"more":[{"key":"latest","count":`,
		`{"key":"top","slugs":[`,
		`"more":[{"key":"top","count":`,
		`{"key":"trending","slugs":[`,
		`"more":[{"key":"trending","count":`,
		`{"key":"updated","slugs":[`,
		`"more":[{"key":"updated","count":`,
		`{"key":"rediscover","slugs":[`,
		`{"key":"core","slugs":["boards"`,
		`"ranked":`,
		`"listings":[{"slug":`,
	} {
		contains(t, home, want)
	}
	for _, gone := range []string{`"title":"New"`, `"note"`, `"empty"`, `"pinned"`, `"activity"`} {
		uassert.False(t, strings.Contains(home, gone), gone+": gnoweb words the front, the store keeps no feed")
	}
	uassert.False(t, strings.Contains(home, `"users"`), "services stay off the store front")
	uassert.Equal(t, notFound, Render("api/v1/nope"))
	uassert.Equal(t, notFound, Render("api/v1/app/blog"), "apps link to their own realm page")
	uassert.Equal(t, notFound, Render("api/v1/category/defi/0"))
	uassert.Equal(t, notFound, Render("api/v1/category/defi"))
	category := Render("api/v1/category/defi/1")
	contains(t, category, `{"version":1,"category":{"key":"defi","label":"DeFi","count":1,"confirmed":1},"page":1`)
	contains(t, category, `"page":1,"pages":1,"listings":[{"slug":`)
	contains(t, Render("api/v1/category/games/1"), `"page":1,"pages":1,"listings":[]`)
}

func TestJSONEscaping(cur realm, t *testing.T) {
	var b strings.Builder
	writeString(&b, "<\"\\>&\x01\u2028é")
	uassert.Equal(t, `"\u003c\"\u005c\u003e\u0026\u0001\u2028é"`, b.String())
	b.Reset()
	writeString(&b, "[Claim](https://example.org)")
	uassert.Equal(t, `"\u005bClaim\u005d(https://example.org)"`, b.String()) // gnoweb renders api/ as markdown: no link

	registerFrom(cur, "gno.land/r/escape/app", "escape-app", `A<"\>B`)
	contains(t, listingJSON("escape-app"), `"title":"A\u003c\"\u005c\u003eB"`)
	var item strings.Builder
	writeItems(&item, []*listing{bySlugOrPanic("escape-app")})
	contains(t, item.String(), `[A\<"\\\>B](/r/escape/app)`)
}

// gnoweb draws $…$, \(…\) and \[…\] as formulas and @name as a link to a
// user: owner text opens none, in the markdown or on the api/ pages, which
// gnoweb renders as markdown too.
func TestNoFormulaFromOwnerText(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/mathns/app"))
	Register(cross(cur), "math-app", "Math $x^2$ App", `Pay \(a\) or \[b\] in $GNOT$, audited by @gnoland`, "utilities", 4)
	l := bySlugOrPanic("math-app")

	var item strings.Builder
	writeItems(&item, []*listing{l})
	contains(t, item.String(), `[Math \$x^2\$ App](/r/mathns/app)`)
	contains(t, item.String(), `in \$GNOT\$, audited by \@gnoland`)

	json := listingJSON("math-app")
	uassert.False(t, strings.ContainsAny(json, "$@") || strings.Contains(json, `\\`), json) // no $ or @, no backslash written as two
	contains(t, json, `"title":"Math \u0024x^2\u0024 App"`)
	contains(t, json, `Pay \u005c(a\u005c) or \u005c\u005bb\u005c\u005d in`)

	// A $ the sanitizer escaped is left alone; one after an escaped
	// backslash is bare.
	for in, want := range map[string]string{
		`a$b`:       `a\$b`,
		`a\$b`:      `a\$b`,
		`a\\$b`:     `a\\\$b`,
		`a\\\$b`:    `a\\\$b`,
		`$`:         `\$`,
		`no dollar`: `no dollar`,
		`a@b`:       `a\@b`,
		`a\@b`:      `a\@b`,
	} {
		uassert.Equal(t, want, inert(in), in)
	}
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

func TestLists(t *testing.T) {
	latest := Render("api/v1/list/latest/1")
	contains(t, latest, `{"version":1,"list":{"key":"latest","count":`)
	contains(t, latest, `"page":1,"pages":`)
	uassert.Equal(t, notFound, Render("api/v1/list/latest/0"))
	uassert.Equal(t, notFound, Render("api/v1/list/latest/999"))
	uassert.Equal(t, notFound, Render("api/v1/list/nope/1"))
	contains(t, Render("api/v1/list/top/1"), `"list":{"key":"top","count":`)
	trend := Render("api/v1/list/trending/1")
	contains(t, trend, `"list":{"key":"trending","count":`)
	contains(t, trend, `"page":1,"pages":1,`)
	uassert.Equal(t, notFound, Render("api/v1/list/trending/2"), "a computed list is one page")
	contains(t, Render("api/v1/list/updated/1"), `"list":{"key":"updated","count":`)

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
		if s.Key != "top" {
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
// apps to gno.land essentials and address namespaces to its list.
func TestFrontSections(t *testing.T) {
	var keys []string
	for _, s := range homeShelves(blockTime()) {
		keys = append(keys, s.Key)
		switch s.Key {
		case "new":
			uassert.True(t, len(s.Items) > 0, "core apps at the head of the index do not stop New")
			for _, l := range s.Items {
				uassert.False(t, l.core, l.Slug+" is a core app")
				uassert.False(t, address(l.Namespace).IsValid(), l.Slug+" is in an address namespace")
			}
		case "core":
			uassert.True(t, len(s.Items) > 0)
		}
	}
	uassert.Equal(t, "new,top,trending,updated,rediscover,core", strings.Join(keys, ","))
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
	uassert.ErrorContains(t, pickPath(path, shown, 7, now), "changed since the proposal")
	uassert.NoError(t, pickPath(path, pickedAs(l, now), 7, now))
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
	uassert.Equal(t, "new", front.Key)
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
// listing never executes against another that took its slug since, and that
// a hide follows its listing to another slug.
func TestProposalBoundToPath(cur realm, t *testing.T) {
	const victim, squatter = "gno.land/r/boundvictim/app", "gno.land/r/boundsquat/app"
	registerFrom(cur, victim, "bound-app", "Bound App")
	ProposeHide(cross(cur), "bound-app", "other")   // checks, then binds to victim
	ProposeSpotlight(cross(cur), "bound-app", 7)    // likewise, and to its title
	uassert.NoError(t, hide("bound-app", "broken")) // another proposal frees the slug
	registerFrom(cur, squatter, "bound-app", "Bound App")

	uassert.ErrorContains(t, hidePath(victim, "other"), "unknown listing")
	uassert.ErrorContains(t, pickPath(victim, pickedAs(bySlugOrPanic("bound-app"), blockTime()), 7, blockTime()), "unknown listing")
	uassert.False(t, bySlugOrPanic("bound-app").hidden)
	uassert.True(t, picked(blockTime()) == nil || picked(blockTime()).Path != squatter)
	uassert.NoError(t, pickPath(squatter, pickedAs(bySlugOrPanic("bound-app"), blockTime()), 1, blockTime()))
	uassert.NoError(t, hidePath(squatter, "other"))

	// A claim that changes slug during the vote is still the one hidden.
	owner := testutils.TestAddress("dodger")
	dodger := "gno.land/r/" + owner.String() + "/scam"
	testing.SetRealm(testing.NewUserRealm(owner))
	Claim(cross(cur), dodger, "dodge-app", "Dodge App", "A tagline", "utilities", 1, kindApp)
	registerFrom(cur, "gno.land/r/dodgehelper/app", "dodge-app", "Dodge Helper") // moves the claim's slug
	uassert.True(t, strings.HasPrefix(lookup(&byPath, dodger).Slug, "dodge-app--"))
	uassert.NoError(t, hidePath(dodger, "phishing"))
	uassert.True(t, lookup(&byPath, dodger).hidden)
	uassert.False(t, bySlugOrPanic("dodge-app").hidden, "the listing that took the slug is not")
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
// :updated but off Recently updated.
func TestFrontLeavesAddressNamespacesOut(cur realm, t *testing.T) {
	setEpoch(5000) // after every other listing
	path := "gno.land/r/" + testutils.TestAddress("freeanon").String() + "/app"
	registerFrom(cur, path, "anon-app", "Anon App")
	skip(updateCooldown)
	testing.SetRealm(testing.NewCodeRealm(path))
	Register(cross(cur), "", "Anon App", "Changed tagline", "utilities", 4)

	l := bySlugOrPanic("anon-app")
	uassert.True(t, apps.updated.Has(descKey(l.Updated, l.Slug)), ":updated lists every app")
	uassert.False(t, strings.Contains(Render(""), "Anon App"), "on no front-page shelf")
	uassert.False(t, strings.Contains(Render("api/v1/home"), `"slug":"anon-app"`))
}

// Every epoch of the trending window is read: a crowd of tied two-star apps
// filed one day cannot hide the week's top app, filed another.
func TestTrendingReadsEverySlot(cur realm, t *testing.T) {
	const e = 5010
	setEpoch(e)
	registerFrom(cur, "gno.land/r/hotslot/app", "hotslot-app", "Hot Slot")
	starFrom(cur, "hotslot-app", "hsa", "hsb", "hsc", "hsd", "hse")
	setEpoch(e + 1) // the newer epoch, read first
	for i := 0; i < trendingScan; i++ {
		ns, fan := "slotfill"+strconv.Itoa(i), "fill"+strconv.Itoa(i/10)
		registerFrom(cur, "gno.land/r/"+ns+"/app", ns, "Slot Fill "+strconv.Itoa(i))
		starFrom(cur, ns, fan+"a", fan+"b")
	}
	top := trending(blockTime(), 10)
	uassert.True(t, len(top.items) > 0 && top.items[0].Slug == "hotslot-app", jsonSlugs(top.items))
}

// A core listing is frozen in code: neither its namespace's owner nor a
// realm deployed at its path can rewrite it.
func TestCoreListingIsFrozen(cur realm, t *testing.T) {
	const path = "gno.land/r/demo/defi/grc20factory"
	testing.SetRealm(testing.NewUserRealm(nameUser(cur, "demo")))
	uassert.AbortsContains(t, cur, "store: core_listing", func() {
		Claim(cross(cur), path, "token-factory", "Free GNOT Airdrop", "A tagline", "defi", 1, kindApp)
	})
	registerIn(cur, path, "token-factory", "Free GNOT Airdrop", "defi")
	l := bySlugOrPanic("token-factory")
	uassert.Equal(t, "Token Factory", l.Title)
	uassert.Equal(t, reasonCore, upsert(newListing(path, "", "Free GNOT Airdrop", "A tagline", "defi", 1, kindApp), blockTime()))

	testing.SetRealm(testing.NewCodeRealm(path))
	uassert.AbortsContains(t, cur, "store: core_listing", func() { SubmitRich(cross(cur), testIcon, "") })
}

// An app's own Register takes its slug from an unconfirmed claim, which
// moves to another one: the app's init runs once, the claim proves nothing.
// The claim itself is never dropped: its listing and its stars stay, so
// there is nothing to gain by listing over it.
func TestRegisterMovesSquattingClaim(cur realm, t *testing.T) {
	squatter := testutils.TestAddress("squatter")
	squat := "gno.land/r/" + squatter.String() + "/squat"
	testing.SetRealm(testing.NewUserRealm(squatter))
	Claim(cross(cur), squat, "real-app", "Real App", "A tagline", "utilities", 1, kindApp)
	starFrom(cur, "real-app", "squatfana", "squatfanb")
	claim := lookup(&byPath, squat)

	registerFrom(cur, "gno.land/r/realns/app", "real-app", "Real App")
	real := bySlugOrPanic("real-app")
	uassert.Equal(t, "gno.land/r/realns/app", real.Path)
	uassert.Equal(t, "real-app", titleHolder(real))
	uassert.True(t, lookup(&byPath, squat) == claim, "the claim is kept")
	moved := claim.Slug
	uassert.True(t, strings.HasPrefix(moved, "real-app--"), moved)
	uassert.True(t, bySlugOrPanic(moved) == claim)
	uassert.False(t, validSlug(moved), "under a slug no listing can ask for, so it is not moved again")
	uassert.Equal(t, 2, claim.starCount(), "with its stars")
	uassert.True(t, categoryMembers("utilities").Has(claim.catKey()), "and its place in its category")

	// With its title held by a confirmed listing, the claim cannot be
	// confirmed under it.
	starFrom(cur, moved, "squatfanc")
	uassert.Equal(t, confirmStars, claim.rankedCount())
	uassert.False(t, claim.Confirmed)
	uassert.Equal(t, "real-app", titleHolder(claim))

	// A moved claim is still its owner's to change, and its realm's to
	// confirm. It keeps the slug it was given while the one it asks for is
	// held or is no slug, and takes it once free, for good.
	testing.SetRealm(testing.NewUserRealm(squatter))
	Claim(cross(cur), squat, "real-app", "Real App Fork", "Another tagline", "utilities", 1, kindApp)
	uassert.Equal(t, "Another tagline", claim.Tagline)
	uassert.True(t, claim.Confirmed, "under a free title, its three stars confirm it")
	uassert.Equal(t, moved, claim.Slug)
	registerIn(cur, squat, "Not A Slug", "Real App Fork", "utilities")
	uassert.True(t, claim.selfListed)
	uassert.Equal(t, moved, claim.Slug)
	registerIn(cur, squat, "real-fork", "Real App Fork", "utilities")
	uassert.Equal(t, "real-fork", claim.Slug)
	uassert.True(t, bySlugOrPanic("real-fork") == claim)
	uassert.False(t, listings.Has(moved))
	uassert.Equal(t, "real-fork", titleHolder(claim), "indexed under its new slug")
	uassert.True(t, lookup(&apps.created, descKey(claim.Created, "real-fork")) == claim)
	registerIn(cur, squat, "real-again", "Real App Fork", "utilities")
	uassert.Equal(t, "real-fork", claim.Slug, "a slug it asked for is kept")

	// A confirmed listing holds its title and its slug.
	registerFrom(cur, "gno.land/r/realcopy/app", "real-copy", "Real App")
	uassert.False(t, listings.Has("real-copy"))
	registerFrom(cur, "gno.land/r/realcopy/app", "real-app", "Real Copy")
	uassert.False(t, byPath.Has("gno.land/r/realcopy/app"))

	// A refused Register moves nothing: a claim keeps its slug when the
	// Register naming it is refused for its title.
	claimed := "gno.land/r/" + squatter.String() + "/other"
	testing.SetRealm(testing.NewUserRealm(squatter))
	Claim(cross(cur), claimed, "kept-claim", "Kept Claim", "A tagline", "utilities", 1, kindApp)
	registerFrom(cur, "gno.land/r/thief/app", "kept-claim", "Boards")
	uassert.Equal(t, "kept-claim", lookup(&byPath, claimed).Slug)
	uassert.False(t, byPath.Has("gno.land/r/thief/app"))

	// The longest slug, ending as a moved one would: the claim still gets a
	// slug of its own, never the one it leaves, and both listings stay
	// reachable by their slugs.
	long := strings.Repeat("a", maxSlug-2) + "-2"
	longClaim := "gno.land/r/" + squatter.String() + "/long"
	testing.SetRealm(testing.NewUserRealm(squatter))
	Claim(cross(cur), longClaim, long, "Long Claim", "A tagline", "utilities", 1, kindApp)
	registerFrom(cur, "gno.land/r/longns/app", long, "Long App")
	lc := lookup(&byPath, longClaim)
	uassert.Equal(t, "gno.land/r/longns/app", bySlugOrPanic(long).Path)
	uassert.True(t, lc.Slug != long && len(lc.Slug) == maxSlug, lc.Slug)
	uassert.True(t, bySlugOrPanic(lc.Slug) == lc)
	uassert.NoError(t, hidePath(longClaim, "other"))
	uassert.True(t, lc.hidden)
	uassert.False(t, bySlugOrPanic(long).hidden, "a hide follows the path to the claim, not the slug to another listing")
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
	uassert.Equal(t, "", titleHolder(l), "and holding no title")

	registerFrom(cur, "gno.land/r/selflisted/app", "self-listed", "Self Listed")
	starFrom(cur, "self-listed", "selffana", "selffanb", "selffanc")
	testing.SetRealm(testing.NewUserRealm(nameUser(cur, "selffana")))
	Unstar(cross(cur), "self-listed")
	uassert.True(t, bySlugOrPanic("self-listed").Confirmed)
}

// A package is only ever a claim. One that lost its confirmation, then its
// slug to a realm, takes it back once that realm is hidden: by its owner's
// next Claim, even one that changes nothing.
func TestMovedClaimTakesItsSlugBack(cur realm, t *testing.T) {
	owner := testutils.TestAddress("jsonkit")
	path := "gno.land/p/" + owner.String() + "/jsonkit"
	testing.SetRealm(testing.NewUserRealm(owner))
	Claim(cross(cur), path, "jsonkit", "JSON Kit", "A tagline", "dev-tools", 1, kindPackage)
	starFrom(cur, "jsonkit", "kitfana", "kitfanb", "kitfanc")
	pkg := lookup(&byPath, path)
	uassert.True(t, pkg.Confirmed)
	testing.SetRealm(testing.NewUserRealm(nameUser(cur, "kitfana")))
	Unstar(cross(cur), "jsonkit")
	registerFrom(cur, "gno.land/r/kitsquat/app", "jsonkit", "Kit Squat")
	moved := pkg.Slug
	uassert.True(t, strings.HasPrefix(moved, "jsonkit--"), moved)

	testing.SetRealm(testing.NewUserRealm(owner))
	Claim(cross(cur), path, "jsonkit", "JSON Kit", "A tagline", "dev-tools", 1, kindPackage)
	uassert.Equal(t, moved, pkg.Slug, "held by a confirmed listing")
	uassert.NoError(t, hide("jsonkit", "impersonation"))
	paused = true
	testing.SetRealm(testing.NewUserRealm(owner))
	uassert.AbortsContains(t, cur, "store: paused", func() {
		Claim(cross(cur), path, "jsonkit", "JSON Kit", "A tagline", "dev-tools", 1, kindPackage)
	})
	paused = false
	uassert.Equal(t, moved, pkg.Slug, "nothing moves during a pause")

	testing.SetRealm(testing.NewUserRealm(owner))
	Claim(cross(cur), path, "jsonkit", "JSON Kit", "A tagline", "dev-tools", 1, kindPackage)
	uassert.Equal(t, "jsonkit", pkg.Slug)
	uassert.True(t, bySlugOrPanic("jsonkit") == pkg)
	uassert.False(t, listings.Has(moved))
	uassert.Equal(t, 2, pkg.starCount(), "with its stars")

	// As a new claim, it takes no slug from another claim; confirmed, it
	// does, and that claim moves in turn.
	other := testutils.TestAddress("kitother")
	otherPath := "gno.land/r/" + other.String() + "/app"
	testing.SetRealm(testing.NewUserRealm(other))
	Claim(cross(cur), otherPath, "kit-two", "Kit Two", "A tagline", "utilities", 1, kindApp)
	registerFrom(cur, "gno.land/r/kitthree/app", "kit-two", "Kit Three") // moves it
	otherClaim := lookup(&byPath, otherPath)
	testing.SetRealm(testing.NewUserRealm(other))
	Claim(cross(cur), otherPath, "jsonkit", "Kit Two", "A tagline", "utilities", 1, kindApp)
	uassert.True(t, strings.Contains(otherClaim.Slug, movedMark), "a claim holds jsonkit")
	registerIn(cur, otherPath, "jsonkit", "Kit Two", "utilities")
	uassert.Equal(t, "jsonkit", otherClaim.Slug, "its realm confirms it: the claim gives way")
	uassert.True(t, strings.HasPrefix(pkg.Slug, "jsonkit--"), pkg.Slug)
	uassert.True(t, bySlugOrPanic(pkg.Slug) == pkg)

	// A hidden listing takes none: its unhide decides.
	uassert.NoError(t, hidePath(path, "other"))
	testing.SetRealm(testing.NewUserRealm(owner))
	Claim(cross(cur), path, "jsonkit", "JSON Kit", "A tagline", "dev-tools", 1, kindPackage)
	uassert.True(t, strings.HasPrefix(pkg.Slug, "jsonkit--"), pkg.Slug)
}

// movedClaim lists a claim of owner o at path under slug and has a realm
// take that slug, so the claim holds a moved one.
func movedClaim(cur realm, t *testing.T, o address, path, slug, title string) *listing {
	t.Helper()
	uassert.Equal(t, "", tryClaim(cur, o, path, slug, title, "utilities", kindApp))
	registerFrom(cur, "gno.land/r/sq"+strings.ReplaceAll(slug, "-", "")+"/app", slug, title+" Squat")
	l := lookup(&byPath, path)
	uassert.True(t, isMovedSlug(l) && bySlugOrPanic(l.Slug) == l, l.Slug)
	return l
}

// claimAgain claims path again, saying what it already says, and asking
// for slug.
func claimAgain(cur realm, o address, path, slug string) string {
	l := lookup(&byPath, path)
	return tryClaim(cur, o, path, slug, l.Title, l.Category, l.Kind)
}

func noBrokenInvariant(t *testing.T, step string) {
	t.Helper()
	for _, m := range checkInvariants(blockTime()) {
		t.Errorf("%s: %s", step, m)
	}
}

// A hidden core listing keeps its slug: a moved listing asking for it, as a
// claim, confirmed by stars or by its realm, keeps its own. (From an
// adversarial pass: slugTaken's core loop comparing the wrong slug passed
// every other test.)
func TestHiddenCoreSlugIsNeverTaken(cur realm, t *testing.T) {
	setEpoch(5500) // after every other star, for the full checks
	core := coreSnapshot()
	o := testutils.TestAddress("hidcoreslug")
	p := "gno.land/r/" + o.String() + "/app"
	l := movedClaim(cur, t, o, p, "hidcore-a", "Hid Core A")
	moved := l.Slug
	cores := []*listing{bySlugOrPanic("boards"), bySlugOrPanic("coins"), bySlugOrPanic("markdown")}
	for _, c := range cores {
		uassert.NoError(t, hidePath(c.Path, "other"))
	}
	ask := func(as string) {
		for _, c := range cores {
			uassert.Equal(t, "", claimAgain(cur, o, p, c.Slug), as)
			uassert.Equal(t, moved, l.Slug, as+" asks for "+c.Slug)
			uassert.False(t, listings.Has(c.Slug), as)
		}
		noBrokenInvariant(t, as)
	}
	ask("a claim")
	starFrom(cur, moved, "hidcorefana", "hidcorefanb", "hidcorefanc")
	uassert.True(t, l.Confirmed)
	ask("a star-confirmed claim")
	registerIn(cur, p, "coins", "Hid Core A", "utilities")
	uassert.Equal(t, moved, l.Slug, "its realm asks for coins")
	for _, c := range cores {
		uassert.NoError(t, unhide(c.Path))
		uassert.True(t, bySlugOrPanic(c.Slug) == c)
	}
	uassert.Equal(t, core, coreSnapshot())
	noBrokenInvariant(t, "cores back")
}

// A hidden moved listing takes no slug and reserves none; unhidden, it takes
// the slug it asks for. (From an adversarial pass: takeSlug without its
// hidden check passed every other test.)
func TestHiddenMovedListing(cur realm, t *testing.T) {
	setEpoch(5500) // after every other star, for the full checks
	o := testutils.TestAddress("hidmoved")
	p := "gno.land/r/" + o.String() + "/app"
	l := movedClaim(cur, t, o, p, "hidmoved-a", "Hid Moved A")
	moved := l.Slug
	uassert.NoError(t, hidePath(p, "other"))
	uassert.Equal(t, "", claimAgain(cur, o, p, "hidmoved-free"))
	uassert.Equal(t, "", tryClaim(cur, o, p, "hidmoved-free", "Hid Moved A", "dev-tools", kindApp))
	registerIn(cur, p, "hidmoved-free", "Hid Moved A", "dev-tools")
	uassert.True(t, l.Confirmed && l.hidden)
	uassert.Equal(t, moved, l.Slug)
	uassert.False(t, listings.Has("hidmoved-free") || listings.Has(moved))
	noBrokenInvariant(t, "hidden")

	registerFrom(cur, "gno.land/r/hidmovedtaker/app", "hidmoved-free", "Hid Moved Taker") // it reserved nothing
	uassert.NoError(t, unhide(p))
	uassert.True(t, bySlugOrPanic(moved) == l)
	registerIn(cur, p, "hidmoved-mine", "Hid Moved A", "dev-tools")
	uassert.Equal(t, "hidmoved-mine", l.Slug)
	uassert.False(t, listings.Has(moved))
	noBrokenInvariant(t, "took a slug after its unhide")
}

// A hidden claim with the stars to be confirmed, and its title held, comes
// back once its owner fixed it while hidden: claiming it again confirms
// nothing. (From an adversarial pass.)
func TestClaimAgainWhileHiddenKeepsItsUnhide(cur realm, t *testing.T) {
	setEpoch(5500) // after every other star, for the full checks
	o := testutils.TestAddress("claimagainhid")
	p := "gno.land/r/" + o.String() + "/app"
	uassert.Equal(t, "", tryClaim(cur, o, p, "again-claim", "Again Hidden", "utilities", kindApp))
	registerFrom(cur, "gno.land/r/againholder/app", "again-holder", "Again Hidden")
	starFrom(cur, "again-claim", "againfana", "againfanb", "againfanc")
	c := lookup(&byPath, p)
	uassert.NoError(t, hide("again-claim", "broken"))
	testing.SetRealm(testing.NewUserRealm(o))
	Claim(cross(cur), p, "again-claim", "Again Hidden", "Fixed now", "utilities", 1, kindApp)
	uassert.False(t, c.Confirmed)
	uassert.NoError(t, unhide(p))
	uassert.False(t, c.Confirmed, "its title is still held")
	noBrokenInvariant(t, "unhidden")
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
	uassert.AbortsContains(t, cur, "a successor is a realm path", func() { ProposeSuccessor(cross(cur), runtime.ChainDomain()+storeURL) }) // never itself
	ProposeSuccessor(cross(cur), "gno.land/r/gnoland/store/v1")

	successor = "gno.land/r/gnoland/store/v1"
	defer func() { successor = "" }()
	moved := "Explore has moved: [/r/gnoland/store/v1](/r/gnoland/store/v1)"
	contains(t, Render(""), moved)
	contains(t, Render("top"), moved)
	contains(t, Block("latest", 1), moved)
	uassert.Equal(t, moved, Badge("gno.land/r/gnoland/blog"))
	contains(t, Render("api/v1/home"), `"moved_to":"gno.land/r/gnoland/store/v1"`)
	contains(t, Render("api/v1/build"), `"moved_to":"gno.land/r/gnoland/store/v1"`)
}

// Only the seed's listings are core: gnoweb's Core badge reads this flag,
// not its trust list.
func TestCoreFlag(t *testing.T) {
	contains(t, listingJSON("blog"), `"core":true`)
	contains(t, listingJSON("starred-app"), `"core":false`)
}

// A shelf cut from a computed list holds gridPool apps, each with its own
// count of the week.
func TestTrendingShelfIsCut(cur realm, t *testing.T) {
	setEpoch(6000) // after every other trending app
	for i := 0; i <= gridPool; i++ {
		ns := "shelfcut" + strconv.Itoa(i)
		registerFrom(cur, "gno.land/r/"+ns+"/app", ns, "Shelf Cut "+strconv.Itoa(i))
		starFrom(cur, ns, "shelfcuta", "shelfcutb")
	}
	s := listShelf(mustList("trending", blockTime()))
	uassert.Equal(t, gridPool, len(s.Items))
	contains(t, listingJSON(s.Items[0].Slug), `"ranked":2,"stars_7d":2,`)
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

// A claim confirmed again after its backers left is not counted again:
// unstarring and starring in a loop does not inflate the pulse.
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
}

// A claim confirmed while hidden is counted, and takes its title, once
// GovDAO unhides it.
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
	uassert.Equal(t, "hidden-claim", titleHolder(l))
}

// A claim holds no title, whoever asks for it: a listed realm may rename
// itself to it, and the claim is left as it is. Its owner then claims it
// again under another title.
func TestRetitleLeavesClaims(cur realm, t *testing.T) {
	owner := testutils.TestAddress("retitleowner")
	lib := "gno.land/p/" + owner.String() + "/lib"
	testing.SetRealm(testing.NewUserRealm(owner))
	Claim(cross(cur), lib, "retitle-lib", "Retitle Lib", "A tagline", "dev-tools", 1, kindPackage)
	registerFrom(cur, "gno.land/r/retitler/app", "retitler", "Retitler")
	registerFrom(cur, "gno.land/r/retitler/app", "", "Retitle Lib")
	uassert.Equal(t, "Retitle Lib", bySlugOrPanic("retitler").Title)
	uassert.Equal(t, "retitler", titleHolder(bySlugOrPanic("retitler")))
	uassert.False(t, byTitle.Has(titleKey("Retitler")), "the old title is released")
	uassert.True(t, bySlugOrPanic("retitle-lib") == lookup(&byPath, lib), "the claim is untouched")
}

// A realm that registers again what its listing already says changes
// nothing, and is not refused for it.
func TestRelistIsIdempotent(cur realm, t *testing.T) {
	const path = "gno.land/r/relistns/app"
	registerFrom(cur, path, "relist-app", "Relist App")
	l := bySlugOrPanic("relist-app")
	created, updated := l.Created, l.Updated
	skip(updateCooldown)
	registerFrom(cur, path, "another-slug", "Relist App")
	uassert.True(t, l == bySlugOrPanic("relist-app") && l.selfListed && l.Confirmed)
	uassert.Equal(t, created, l.Created)
	uassert.Equal(t, updated, l.Updated)

	// A claimed realm that registers what its claim says is confirmed by it.
	owner := testutils.TestAddress("relistclaim")
	claimed := "gno.land/r/" + owner.String() + "/app"
	testing.SetRealm(testing.NewUserRealm(owner))
	Claim(cross(cur), claimed, "relist-claim", "Relist Claim", "A tagline", "utilities", 4, kindApp)
	registerFrom(cur, claimed, "", "Relist Claim")
	c := bySlugOrPanic("relist-claim")
	uassert.True(t, c.Confirmed && c.selfListed)
	uassert.Equal(t, "relist-claim", titleHolder(c))
}

// titleKey's fast path for ASCII gives the key its general path would.
func TestTitleKeyASCII(t *testing.T) {
	for r := rune(0); r < utf8.RuneSelf; r++ {
		want := ""
		if c := foldRune(unicode.ToLower(r)); c == '"' {
			want = "''" // as titleKey writes a double quote
		} else if c != 0 {
			want = string(c)
		}
		uassert.Equal(t, want, titleKey(string(r)), strconv.Itoa(int(r)))
	}
	uassert.Equal(t, titleKey("Modem"), titleKey("MODERN"), "rn and m still fold")
}

// A hidden listing shares its slug with the listing that took it, and takes
// no title from it by registering again.
func TestHiddenListingTakesNoTitle(cur realm, t *testing.T) {
	registerFrom(cur, "gno.land/r/hidfirst/app", "hid-app", "Hid First")
	uassert.NoError(t, hide("hid-app", "other"))
	registerFrom(cur, "gno.land/r/hidreal/app", "hid-app", "Hid Real")
	registerFrom(cur, "gno.land/r/hidfirst/app", "", "Hid Real")
	uassert.Equal(t, "Hid First", lookup(&byPath, "gno.land/r/hidfirst/app").Title)

	// Its image can still change while hidden, and puts it on no shelf.
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/hidfirst/app"))
	SubmitRich(cross(cur), testIcon, "")
	hidden := lookup(&byPath, "gno.land/r/hidfirst/app")
	uassert.True(t, hidden.hidden && hidden.next != nil)
	uassert.False(t, lookup(&apps.created, descKey(hidden.Created, hidden.Slug)) == hidden)
}

// A core listing can always be unhidden, even from a namespace that is not
// a core one and filled up meanwhile; a claim whose title was taken when
// its stars came is confirmed by them at its unhide.
func TestUnhide(cur realm, t *testing.T) {
	const md = "gno.land/p/moul/md/v0"
	uassert.NoError(t, hidePath(md, "broken"))
	for i := 0; visibleInNamespace("moul") < maxPerNamespace; i++ {
		registerFrom(cur, "gno.land/r/moul/fill"+strconv.Itoa(i), "moul-fill-"+strconv.Itoa(i), "Moul Fill "+strconv.Itoa(i))
	}
	uassert.NoError(t, unhide(md))
	uassert.True(t, bySlugOrPanic("markdown").core)

	owner := testutils.TestAddress("unhideclaim")
	path := "gno.land/p/" + owner.String() + "/lib"
	testing.SetRealm(testing.NewUserRealm(owner))
	Claim(cross(cur), path, "unhide-claim", "Unhide Title", "A tagline", "dev-tools", 1, kindPackage)
	registerFrom(cur, "gno.land/r/unhideholder/app", "unhide-holder", "Unhide Title") // a realm's first listing shares it
	starFrom(cur, "unhide-claim", "unhidefana", "unhidefanb", "unhidefanc")
	claim := bySlugOrPanic("unhide-claim")
	uassert.False(t, claim.Confirmed, "its title is held")
	uassert.NoError(t, hide("unhide-claim", "other"))
	registerFrom(cur, "gno.land/r/unhideholder/app", "", "Unhide Holder") // frees the title
	uassert.NoError(t, unhide(path))
	uassert.True(t, claim.Confirmed)
	uassert.Equal(t, "unhide-claim", titleHolder(claim))

	// A claim held no title before its hide: registering it meanwhile keeps
	// no one from unhiding the claim, which stays one. Its realm's listing,
	// confirmed while hidden, would hold the title, so that one waits.
	shared := "gno.land/r/" + owner.String() + "/shared"
	testing.SetRealm(testing.NewUserRealm(owner))
	Claim(cross(cur), shared, "shared-claim", "Shared Title", "A tagline", "utilities", 1, kindApp)
	uassert.NoError(t, hide("shared-claim", "other"))
	registerFrom(cur, "gno.land/r/sharedholder/app", "shared-holder", "Shared Title")
	uassert.NoError(t, unhide(shared))
	sc := bySlugOrPanic("shared-claim")
	uassert.False(t, sc.Confirmed)
	uassert.Equal(t, "shared-holder", titleHolder(sc))
	testing.SetRealm(testing.NewUserRealm(owner))
	uassert.AbortsContains(t, cur, "store: title_taken", func() { // a new claim still takes no held title
		Claim(cross(cur), "gno.land/r/"+owner.String()+"/sharedtwo", "shared-two", "Shared Title", "A tagline", "utilities", 1, kindApp)
	})

	uassert.NoError(t, hide("shared-claim", "other"))
	registerIn(cur, shared, "", "Shared Title", "utilities") // refused: the title changes hands
	uassert.False(t, sc.Confirmed)
	registerFrom(cur, "gno.land/r/sharedholder/app", "", "Shared Holder") // frees the title
	registerIn(cur, shared, "", "Shared Title", "utilities")
	uassert.True(t, sc.Confirmed && sc.hidden)
	registerFrom(cur, "gno.land/r/sharedlate/app", "shared-late", "Shared Title") // takes it while the listing is hidden
	uassert.ErrorContains(t, unhide(shared), reasonTitleTaken)
}

// A refused Claim names the confirmed listing holding the title: the owner
// learns what to ask GovDAO to hide, if it is a squat.
func TestClaimNamesHolder(cur realm, t *testing.T) {
	owner := testutils.TestAddress("uuidowner")
	testing.SetRealm(testing.NewUserRealm(owner))
	uassert.AbortsContains(t, cur, "store: title_taken (holder: boards)", func() {
		Claim(cross(cur), "gno.land/p/"+owner.String()+"/boards", "boards-lib", "Boards", "A tagline", "dev-tools", 1, kindPackage)
	})
}

// Two claims may share a title, so a claim squats none: the first confirmed
// takes it.
func TestClaimsShareATitle(cur realm, t *testing.T) {
	squatter, owner := testutils.TestAddress("uuidsquatter"), testutils.TestAddress("uuidrealowner")
	testing.SetRealm(testing.NewUserRealm(squatter))
	Claim(cross(cur), "gno.land/p/"+squatter.String()+"/uuid", "uuid-squat", "UUID", "A tagline", "dev-tools", 1, kindPackage)
	testing.SetRealm(testing.NewUserRealm(owner))
	Claim(cross(cur), "gno.land/p/"+owner.String()+"/uuid", "uuid", "UUID", "A tagline", "dev-tools", 1, kindPackage)

	starFrom(cur, "uuid", "uuidfana", "uuidfanb", "uuidfanc")
	uassert.True(t, bySlugOrPanic("uuid").Confirmed)
	uassert.Equal(t, "uuid", titleHolder(bySlugOrPanic("uuid")))
	starFrom(cur, "uuid-squat", "uuidfana", "uuidfanb", "uuidfanc")
	other := bySlugOrPanic("uuid-squat")
	uassert.False(t, other.Confirmed, "the title is held")

	// Claimed again under a free title, it is confirmed by the stars it has.
	testing.SetRealm(testing.NewUserRealm(squatter))
	Claim(cross(cur), other.Path, "", "UUID Too", "A tagline", "dev-tools", 1, kindPackage)
	uassert.True(t, other.Confirmed)
	uassert.Equal(t, "uuid-squat", titleHolder(other))
}

// A path longer than any package's is refused: every read would carry it.
func TestPathLength(cur realm, t *testing.T) {
	owner := testutils.TestAddress("longpath")
	base := "gno.land/r/" + owner.String() + "/"
	fits := base + strings.Repeat("a", maxPath-len(base))
	uassert.Equal(t, owner.String(), namespaceOf(fits))
	uassert.Equal(t, "", namespaceOf(fits+"a"))
	testing.SetRealm(testing.NewUserRealm(owner))
	uassert.AbortsContains(t, cur, "does not own the namespace", func() {
		Claim(cross(cur), fits+"a", "long-path", "Long Path", "A tagline", "utilities", 1, kindApp)
	})
}

// A category files its confirmed apps before its claims, and counts them
// apart: claims never push a confirmed app down its page.
func TestCategoryFilesClaimsLast(cur realm, t *testing.T) {
	registerIn(cur, "gno.land/r/gamesns/app", "games-app", "Games App", "games")
	confirmed := confirmedCount("games")
	for i := 0; i < 3; i++ { // newer than the confirmed app
		skip(1)
		owner := testutils.TestAddress("gameclaim" + strconv.Itoa(i))
		testing.SetRealm(testing.NewUserRealm(owner))
		Claim(cross(cur), "gno.land/r/"+owner.String()+"/game", "game-claim-"+strconv.Itoa(i), "Game Claim", "A tagline", "games", 1, kindApp)
	}
	uassert.Equal(t, confirmed, confirmedCount("games"))
	uassert.Equal(t, confirmed+3, appCount("games"))
	first, _ := pageOf(categoryMembers("games"), 1)
	for i, l := range first {
		uassert.Equal(t, i < confirmed, l.Confirmed, l.Slug)
	}
	contains(t, Render("api/v1/category/games/1"), `"count":`+strconv.Itoa(confirmed+3)+`,"confirmed":`+strconv.Itoa(confirmed)+`}`)
}

// The typed reads return copies of every listing, hidden ones included, and
// of its stars, a bounded page at a time.
func TestReads(cur realm, t *testing.T) {
	registerFrom(cur, "gno.land/r/readsns/app", "reads-app", "Reads App")
	starFrom(cur, "reads-app", "readsfana")
	starAnon(cur, "reads-app", "readsanon")

	l, ok := ListingOf("gno.land/r/readsns/app")
	uassert.True(t, ok)
	uassert.Equal(t, "reads-app", l.Slug)
	uassert.True(t, l.Confirmed && !l.Hidden && !l.Core)
	uassert.Equal(t, 2, l.Stars)
	uassert.Equal(t, 1, l.Ranked)
	_, ok = ListingOf("gno.land/r/readsns/nope")
	uassert.False(t, ok)

	stars := StarsOf("gno.land/r/readsns/app", 0, 10)
	uassert.Equal(t, 2, len(stars))
	ranked := 0
	for _, s := range stars {
		uassert.True(t, s.Address.IsValid())
		if s.Ranked {
			ranked++
		}
	}
	uassert.Equal(t, 1, ranked)
	uassert.Equal(t, 1, len(StarsOf("gno.land/r/readsns/app", 1, 10)))
	uassert.Equal(t, 0, len(StarsOf("gno.land/r/readsns/nope", 0, 10)))

	uassert.NoError(t, hide("reads-app", "other"))
	l, _ = ListingOf("gno.land/r/readsns/app")
	uassert.True(t, l.Hidden, "a hidden listing is still read")

	// Paging through Listings reads every listing once.
	seen := 0
	for offset := 0; ; offset += 7 {
		page := Listings(offset, 7)
		seen += len(page)
		if len(page) < 7 {
			break
		}
	}
	uassert.Equal(t, ListingCount(), seen)
	uassert.Equal(t, readPage, len(Listings(0, 1<<30)), "a read is bounded")
	uassert.Equal(t, 0, len(Listings(0, -1)))
	uassert.False(t, Paused())
	uassert.Equal(t, "", Successor())
}

// A response's buffer grows by doubling. Left to grow as it is written, it
// would be copied whole at every write, and a full front page would cost
// more gas than a query may spend.
func TestResponseGrowsByDoubling(t *testing.T) {
	var b strings.Builder
	l, now := bySlugOrPanic("blog"), blockTime()
	grown, last := 0, 0
	writeArray(&b, 100, func(int) {
		writeListing(&b, l, now)
		if c := b.Cap(); c != last {
			grown, last = grown+1, c
		}
	})
	uassert.True(t, grown <= 8, "grown "+strconv.Itoa(grown)+" times for 100 listings")
}

// Random sequences of every write, with every index rebuilt and compared
// as they go. They run last: they move the clock far ahead.
func TestRandomWrites(cur realm, t *testing.T) {
	did := map[string]int{}
	for _, run := range []map[string]int{
		runRandom(cur, t, "a", 1, 150, 7100),
		runRandom(cur, t, "b", 0xdecafbad, 150, 9000),
		runRandom(cur, t, "c", 31337, 150, 11000),
	} {
		for what, n := range run {
			did[what] += n
		}
	}
	// Rarer than what each run must do (runRandom), and set by what the
	// suite left before: once in all is enough.
	uassert.True(t, did["reslug"] > 0, "no claim was moved")
	uassert.True(t, did["slug back"] > 0, "no moved listing took a slug back")
	uassert.True(t, did["unhide moved"] > 0, "no listing came back from a hide under a moved slug")
}

// TestInvariants checks what the whole suite left: each index holds exactly
// what the listings say, and every response is valid JSON.
func TestInvariants(t *testing.T) {
	setEpoch(13000) // after the random runs: every test starts from the default clock
	for _, m := range append(checkInvariants(blockTime()), checkJSON()...) {
		t.Errorf("%s", m)
	}
}
