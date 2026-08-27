package grc721reg

import (
	"strings"
	"testing"

	"gno.land/p/demo/tokens/grc721"
	"gno.land/p/demo/tokens/grc721/collection"
	"gno.land/p/demo/tokens/grc721/enumerable"
	"gno.land/p/demo/tokens/grc721/metadata"
	"gno.land/p/demo/tokens/grc721/royalty"
	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"
)

var (
	alice  = testutils.TestAddress("alice")
	artist = testutils.TestAddress("artist")
)

// Must run before any Register in this file (shared registry; empty home path).
func TestRenderEmptyRegistry(cur realm, t *testing.T) {
	urequire.Equal(t, "No registered collection.", Render(""),
		"empty registry renders the placeholder line")
}

func TestRegisterAndLookup(cur realm, t *testing.T) {
	type regCase struct {
		name      string
		realmPath string
		symbol    string
		tokenName string
		slug      string
		build     func(cur realm) *collection.Collection
		wantKinds []string
		verify    func(t *testing.T, key string)
	}

	cases := []regCase{
		{
			name:      "register core-only collection",
			realmPath: "gno.land/r/demo/reg_core",
			symbol:    "CORE",
			tokenName: "Core Only",
			slug:      "",
			build: func(cur realm) *collection.Collection {
				token, _ := grc721.NewToken("Core Only", "CORE", 0, cur)
				return collection.NewCollection(token)
			},
			wantKinds: []string{},
		},
		{
			name:      "register royalty-only collection with a valid slug",
			realmPath: "gno.land/r/demo/reg_roy",
			symbol:    "ROY",
			tokenName: "Royalty Only",
			slug:      "My-Cool_Slug123",
			build: func(cur realm) *collection.Collection {
				token, ledger := grc721.NewToken("Royalty Only", "ROY", 0, cur)
				roy, royLed := royalty.NewRoyalty(ledger, 1000)
				urequire.NoError(t, ledger.Mint(alice, "1"))
				urequire.NoError(t, royLed.SetTokenRoyalty("1", artist, 250))
				return collection.NewCollection(token, roy)
			},
			wantKinds: []string{royalty.Kind},
			verify: func(t *testing.T, key string) {
				v, ok := GetExtension(key, royalty.Kind)
				urequire.True(t, ok, "royalty present")
				_, amount, err := v.(*royalty.Royalty).RoyaltyInfo("1", 10000)
				urequire.NoError(t, err)
				urequire.Equal(t, int64(250), amount)
			},
		},
		{
			name:      "register metadata+royalty+enumerable and look up each by kind",
			realmPath: "gno.land/r/demo/reg_full",
			symbol:    "FULL",
			tokenName: "Full Stack",
			slug:      "full",
			build: func(cur realm) *collection.Collection {
				token, ledger := grc721.NewToken("Full Stack", "FULL", 0, cur)
				meta, metaLed := metadata.NewMetadata(ledger)
				roy, royLed := royalty.NewRoyalty(ledger, 1000)
				enum, _ := enumerable.NewEnumerable(ledger)
				urequire.NoError(t, ledger.Mint(alice, "1"))
				urequire.NoError(t, metaLed.SetTokenURI("1", "ipfs://uri-1"))
				urequire.NoError(t, royLed.SetDefaultRoyalty(artist, 500))
				return collection.NewCollection(token, meta, roy, enum)
			},
			wantKinds: []string{enumerable.Kind, metadata.Kind, royalty.Kind},
			verify: func(t *testing.T, key string) {
				vRoy, ok := GetExtension(key, royalty.Kind)
				urequire.True(t, ok, "royalty present")
				recv, amount, err := vRoy.(*royalty.Royalty).RoyaltyInfo("1", 10000)
				urequire.NoError(t, err)
				urequire.Equal(t, artist, recv)
				urequire.Equal(t, int64(500), amount)

				vMeta, ok := GetExtension(key, metadata.Kind)
				urequire.True(t, ok, "metadata present")
				uri, err := vMeta.(*metadata.Metadata).TokenURI("1")
				urequire.NoError(t, err)
				urequire.Equal(t, "ipfs://uri-1", uri)

				vEnum, ok := GetExtension(key, enumerable.Kind)
				urequire.True(t, ok, "enumerable present")
				urequire.Equal(t, int64(1), vEnum.(*enumerable.Enumerable).TotalSupply())
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			testing.SetRealm(testing.NewCodeRealm(tc.realmPath))
			col := tc.build(cur)
			key := Register(cross(cur), col, tc.slug)

			urequire.NotEqual(t, "", key, "Register returns the fqname key")
			got := Get(key)
			urequire.True(t, got != nil, "Get resolves the registered key")
			urequire.Equal(t, tc.symbol, MustGet(key).Token().GetSymbol(),
				"MustGet resolves the same symbol")

			tok := GetToken(key)
			urequire.True(t, tok != nil, "GetToken resolves the core token")
			urequire.Equal(t, tc.tokenName, tok.GetName(), "core token name matches")

			kinds := MustGet(key).Kinds()
			urequire.Equal(t, len(tc.wantKinds), len(kinds), "kind count matches")
			for i, k := range tc.wantKinds {
				urequire.Equal(t, k, kinds[i], "kind ordering matches")
				urequire.True(t, MustGet(key).HasExtension(k), "HasExtension true for "+k)
			}
			urequire.False(t, MustGet(key).HasExtension("does-not-exist"),
				"HasExtension false for an unattached kind")

			urequire.True(t, GetRegistry().Has(key), "GetRegistry sees the entry")

			if tc.verify != nil {
				tc.verify(t, key)
			}
		})
	}
}

func TestGettersOnAbsentKey(cur realm, t *testing.T) {
	const absent = "gno.land/r/demo/reg_absent.NONE"

	type absentCase struct {
		name  string
		check func(t *testing.T)
	}

	cases := []absentCase{
		{
			name: "Get returns nil for an unknown key",
			check: func(t *testing.T) {
				urequire.True(t, Get(absent) == nil, "Get returns nil")
			},
		},
		{
			name: "GetToken returns nil for an unknown key",
			check: func(t *testing.T) {
				urequire.True(t, GetToken(absent) == nil, "GetToken returns nil")
			},
		},
		{
			name: "GetExtension returns (nil, false) for an unknown key",
			check: func(t *testing.T) {
				v, ok := GetExtension(absent, royalty.Kind)
				urequire.False(t, ok, "ok is false")
				urequire.True(t, v == nil, "view is nil")
			},
		},
		{
			name: "MustGet unknown key aborts",
			check: func(t *testing.T) {
				uassert.PanicsWithMessage(t, cur,
					"grc721reg: unknown collection: "+absent, func() {
						MustGet(absent)
					})
			},
		},
		{
			// Render's input is attacker-controlled, so an unknown key must
			// render a message rather than abort: aborting hands any visitor a
			// way to break the page with a bad URL.
			name: "Render of an unknown key reports it instead of aborting",
			check: func(t *testing.T) {
				out := Render(absent)
				urequire.True(t, strings.Contains(out, "Unknown collection"),
					"renders a not-found message, got: "+out)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.check(t)
		})
	}
}

// Register panics surface as cross-realm aborts (AbortsContains, not PanicsWithMessage).
func TestRegisterAborts(cur realm, t *testing.T) {
	type abortCase struct {
		name   string
		substr string
		run    func(cur realm, t *testing.T)
	}

	cases := []abortCase{
		{
			name:   "register nil collection aborts",
			substr: "nil collection",
			run: func(cur realm, t *testing.T) {
				testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/reg_nil"))
				urequire.AbortsContains(t, cur, "nil collection", func() {
					Register(cross(cur), nil, "")
				})
			},
		},
		{
			name:   "register token from a different realm aborts",
			substr: "must be registered from its own realm",
			run: func(cur realm, t *testing.T) {
				testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/reg_src"))
				token, _ := grc721.NewToken("Mismatch", "MIS", 0, cur)
				col := collection.NewCollection(token)

				testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/reg_dst"))
				urequire.AbortsContains(t, cur, "must be registered from its own realm", func() {
					Register(cross(cur), col, "")
				})
			},
		},
		{
			name:   "duplicate key (same realm and symbol) aborts",
			substr: "collection already registered",
			run: func(cur realm, t *testing.T) {
				testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/reg_dupe"))
				token, _ := grc721.NewToken("Bar", "BAR", 0, cur)
				Register(cross(cur), collection.NewCollection(token), "dup")

				replacement, _ := grc721.NewToken("Replacement", "BAR", 1, cur)
				urequire.AbortsContains(t, cur, "collection already registered", func() {
					Register(cross(cur), collection.NewCollection(replacement), "other")
				})
			},
		},
		{
			name:   "validateSlug rejects a too-long slug",
			substr: "slug too long",
			run: func(cur realm, t *testing.T) {
				testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/reg_slug_long"))
				token, _ := grc721.NewToken("Long", "LONG", 0, cur)
				longSlug := strings.Repeat("a", maxSlugLen+1)
				urequire.AbortsContains(t, cur, "slug too long", func() {
					Register(cross(cur), collection.NewCollection(token), longSlug)
				})
			},
		},
		{
			name:   "validateSlug rejects an invalid slug character",
			substr: "invalid slug character",
			run: func(cur realm, t *testing.T) {
				testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/reg_slug_bad"))
				token, _ := grc721.NewToken("Bad", "BAD", 0, cur)
				urequire.AbortsContains(t, cur, "invalid slug character: /", func() {
					Register(cross(cur), collection.NewCollection(token), "bad/slug")
				})
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.run(cur, t)
		})
	}
}

func TestValidateSlug(cur realm, t *testing.T) {
	type slugCase struct {
		name    string
		slug    string
		wantMsg string // "" = accept
	}

	cases := []slugCase{
		{name: "lowercase, uppercase, digit, underscore and dash accepted",
			slug: "aZ_9-Xy0", wantMsg: ""},
		{name: "empty slug accepted", slug: "", wantMsg: ""},
		{name: "max-length slug accepted", slug: strings.Repeat("a", maxSlugLen), wantMsg: ""},
		{name: "too-long slug rejected",
			slug: strings.Repeat("a", maxSlugLen+1), wantMsg: "grc721reg: slug too long"},
		{name: "slash character rejected",
			slug: "a/b", wantMsg: "grc721reg: invalid slug character: /"},
		{name: "space character rejected",
			slug: "a b", wantMsg: "grc721reg: invalid slug character:  "},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.wantMsg == "" {
				uassert.NotPanics(t, cur, func() {
					validateSlug(tc.slug)
				})
				return
			}
			uassert.PanicsWithMessage(t, cur, tc.wantMsg, func() {
				validateSlug(tc.slug)
			})
		})
	}
}

func TestIsAlphanumeric(cur realm, t *testing.T) {
	type runeCase struct {
		name string
		c    rune
		want bool
	}

	cases := []runeCase{
		{"lowercase a", 'a', true},
		{"lowercase z", 'z', true},
		{"uppercase A", 'A', true},
		{"uppercase Z", 'Z', true},
		{"digit 0", '0', true},
		{"digit 9", '9', true},
		{"underscore is not alphanumeric", '_', false},
		{"dash is not alphanumeric", '-', false},
		{"slash is not alphanumeric", '/', false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			urequire.Equal(t, tc.want, isAlphanumeric(tc.c), tc.name)
		})
	}
}

func TestNewCollectionRejects(cur realm, t *testing.T) {
	type ncCase struct {
		name  string
		msg   string
		build func(cur realm)
	}

	cases := []ncCase{
		{
			name: "nil token rejected",
			msg:  "collection: nil token",
			build: func(cur realm) {
				collection.NewCollection(nil)
			},
		},
		{
			name: "view from a different token rejected",
			msg:  "collection: view does not belong to this token",
			build: func(cur realm) {
				tokenX, _ := grc721.NewToken("X", "X", 0, cur)
				_, ledgerY := grc721.NewToken("Y", "Y", 1, cur)
				royY, _ := royalty.NewRoyalty(ledgerY, 1000)
				collection.NewCollection(tokenX, royY)
			},
		},
		{
			name: "duplicate extension kind rejected",
			msg:  "collection: duplicate extension kind: royalty",
			build: func(cur realm) {
				token, ledger := grc721.NewToken("Dup", "DUP", 0, cur)
				roy, _ := royalty.NewRoyalty(ledger, 1000)
				collection.NewCollection(token, roy, roy)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/reg_newcoll"))
			uassert.PanicsWithMessage(t, cur, tc.msg, func() {
				tc.build(cur)
			})
		})
	}
}

func TestRenderCollections(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/render_core"))
	coreTok, _ := grc721.NewToken("Render Core", "RCORE", 0, cur)
	coreKey := Register(cross(cur), collection.NewCollection(coreTok), "core")

	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/render_ext"))
	extTok, extLed := grc721.NewToken("Render Ext", "REXT", 0, cur)
	enum, _ := enumerable.NewEnumerable(extLed)
	urequire.NoError(t, extLed.Mint(alice, "1"))
	extKey := Register(cross(cur), collection.NewCollection(extTok, enum), "ext")

	type renderCase struct {
		name     string
		path     string
		contains []string
	}

	cases := []renderCase{
		{
			name:     "Render home lists collections with badges",
			path:     "",
			contains: []string{"Render Core", "_core only_", "Render Ext", "`enumerable`", "info"},
		},
		{
			name:     "Render specific core-only key shows core-only badge",
			path:     coreKey,
			contains: []string{"# Render Core", "symbol:", "RCORE", "total supply:", "_core only_"},
		},
		{
			name:     "Render specific key with extension shows the kind badge",
			path:     extKey,
			contains: []string{"# Render Ext", "REXT", "total supply: 1", "`enumerable`"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := Render(tc.path)
			for _, want := range tc.contains {
				urequire.True(t, strings.Contains(out, want),
					"render output should contain "+want)
			}
		})
	}
}
