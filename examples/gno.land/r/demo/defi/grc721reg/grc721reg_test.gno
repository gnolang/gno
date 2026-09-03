package grc721reg

import (
	"chain"
	"strconv"
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

// stubView is an ExtensionView whose kind the registrant picks, which is what
// the registry accepts from any realm it does not import.
type stubView struct{ kind, tokenID string }

func (v *stubView) ExtensionKind() string                       { return v.kind }
func (v *stubView) TokenID() string                             { return v.tokenID }
func (v *stubView) OnMint(address, grc721.TokenID)              {}
func (v *stubView) OnTransfer(address, address, grc721.TokenID) {}
func (v *stubView) OnBurn(grc721.TokenID)                       {}

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
				nft, _ := collection.New("Core Only", "CORE", 0, cur)
				return nft.Collection()
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
				nft, ledger := collection.New("Royalty Only", "ROY", 0, cur)
				_, royLed := royalty.NewRoyalty(ledger, 1000)
				urequire.NoError(t, ledger.Mint(alice, "1"))
				urequire.NoError(t, royLed.SetTokenRoyalty("1", artist, 250))
				return nft.Collection()
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
				nft, ledger := collection.New("Full Stack", "FULL", 0, cur)
				_, metaLed := metadata.NewMetadata(ledger)
				_, royLed := royalty.NewRoyalty(ledger, 1000)
				enumerable.NewEnumerable(ledger)
				urequire.NoError(t, ledger.Mint(alice, "1"))
				urequire.NoError(t, metaLed.SetTokenURI("1", "ipfs://uri-1"))
				urequire.NoError(t, royLed.SetDefaultRoyalty(artist, 500))
				return nft.Collection()
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
		// SetRealm inside the subtest closure does not reach the cur the
		// closure captured, so each case registers from its realm out here.
		testing.SetRealm(testing.NewCodeRealm(tc.realmPath))
		col := tc.build(cur)
		key := Register(cross(cur), col, tc.slug)

		t.Run(tc.name, func(t *testing.T) {
			urequire.Equal(t, tc.realmPath+"."+tc.symbol, key,
				"Register keys the entry on the case's own realm and symbol")
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
		name string
		run  func(cur realm, t *testing.T)
	}

	cases := []abortCase{
		{
			name: "register nil collection aborts",
			run: func(cur realm, t *testing.T) {
				testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/reg_nil"))
				urequire.AbortsContains(t, cur, "nil collection", func() {
					Register(cross(cur), nil, "")
				})
			},
		},
		{
			name: "register collection without token aborts",
			run: func(cur realm, t *testing.T) {
				testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/reg_nil_token"))
				urequire.AbortsContains(t, cur, "grc721reg: nil token", func() {
					Register(cross(cur), &collection.Collection{}, "")
				})
			},
		},
		{
			name: "register token from a different realm aborts",
			run: func(cur realm, t *testing.T) {
				testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/reg_src"))
				_, ledger := grc721.NewToken("Mismatch", "MIS", 0, cur)
				col := collection.NewCollection(ledger)

				testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/reg_dst"))
				urequire.AbortsContains(t, cur, "must be registered from its own realm", func() {
					Register(cross(cur), col, "")
				})
			},
		},
		{
			name: "registration requires a persistent realm",
			run: func(cur realm, t *testing.T) {
				testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/reg_persistent"))
				_, ledger := grc721.NewToken("Persistent", "PERSIST", 0, cur)
				col := collection.NewCollection(ledger)

				testing.SetRealm(testing.NewUserRealm(testutils.TestAddress("runner")))
				urequire.AbortsContains(t, cur, "registration requires a persistent realm", func() {
					Register(cross(cur), col, "")
				})
			},
		},
		{
			name: "duplicate key (same realm and symbol) aborts",
			run: func(cur realm, t *testing.T) {
				testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/reg_dupe"))
				_, ledger := grc721.NewToken("Bar", "BAR", 0, cur)
				Register(cross(cur), collection.NewCollection(ledger), "dup")

				_, replacement := grc721.NewToken("Replacement", "BAR", 1, cur)
				urequire.AbortsContains(t, cur, "collection already registered", func() {
					Register(cross(cur), collection.NewCollection(replacement), "other")
				})
			},
		},
		{
			name: "extension kind longer than the event allows aborts",
			run: func(cur realm, t *testing.T) {
				testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/reg_kind_long"))
				token, ledger := grc721.NewToken("Kind", "KIND", 0, cur)
				view := &stubView{strings.Repeat("k", maxKindLen+1), token.ID()}
				ledger.RegisterExtension(view, view)
				urequire.AbortsContains(t, cur, "extension kind too long", func() {
					Register(cross(cur), collection.NewCollection(ledger), "")
				})
				urequire.True(t, Get("gno.land/r/demo/reg_kind_long.KIND") == nil,
					"a rejected collection is not registered")
			},
		},
		{
			name: "validateSlug rejects a too-long slug",
			run: func(cur realm, t *testing.T) {
				testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/reg_slug_long"))
				_, ledger := grc721.NewToken("Long", "LONG", 0, cur)
				longSlug := strings.Repeat("a", maxSlugLen+1)
				urequire.AbortsContains(t, cur, "slug too long", func() {
					Register(cross(cur), collection.NewCollection(ledger), longSlug)
				})
			},
		},
		{
			name: "validateSlug rejects an invalid slug character",
			run: func(cur realm, t *testing.T) {
				testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/reg_slug_bad"))
				_, ledger := grc721.NewToken("Bad", "BAD", 0, cur)
				urequire.AbortsContains(t, cur, "invalid slug character: /", func() {
					Register(cross(cur), collection.NewCollection(ledger), "bad/slug")
				})
			},
		},
	}

	for _, tc := range cases {
		t.Log(tc.name)
		tc.run(cur, t)
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

func TestRenderCollections(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/render_core"))
	_, coreLed := grc721.NewToken("Render Core", "RCORE", 0, cur)
	coreKey := Register(cross(cur), collection.NewCollection(coreLed), "core")

	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/render_ext"))
	_, extLed := grc721.NewToken("Render Ext", "REXT", 0, cur)
	enumerable.NewEnumerable(extLed)
	urequire.NoError(t, extLed.Mint(alice, "1"))
	extKey := Register(cross(cur), collection.NewCollection(extLed), "ext")

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

func TestRenderEscapesExtensionKinds(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/reg_escape"))
	token, ledger := grc721.NewToken("Esc", "ESC", 0, cur)
	key := Register(cross(cur), collection.NewCollection(ledger), "esc")

	cases := []struct {
		kind string
		want string
	}{
		{"a`b", "``a`b``"},
		{"```", "```` ``` ````"},
		{"line\n# heading", "`line # heading`"},
		{"line\rreturn", "`line return`"},
		{"](http://evil)", "`](http://evil)`"},
		{"# heading", "`# heading`"},
		{"*bold*", "`*bold*`"},
	}

	for _, tc := range cases {
		view := &stubView{tc.kind, token.ID()}
		ledger.RegisterExtension(view, view)
	}

	for _, out := range []string{Render(""), Render(key)} {
		for _, tc := range cases {
			urequire.True(t, strings.Contains(out, tc.want),
				"rendered kind should stay inside its code span: "+tc.kind)
		}
		urequire.False(t, strings.Contains(out, "\n# heading"),
			"a kind cannot open a heading")
		urequire.False(t, strings.Contains(out, "\r"),
			"a kind cannot retain a carriage return")
	}
}

func TestOneTokenTakesOneKey(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/reg_alias"))
	_, ledger := grc721.NewToken("Alias", "ALIAS", 0, cur)
	Register(cross(cur), collection.NewCollection(ledger), "a")

	urequire.AbortsContains(t, cur, "already registered", func() {
		Register(cross(cur), collection.NewCollection(ledger), "b")
	})
}

func TestRegistryLookupGrantsNoTransferAuthority(cur realm, t *testing.T) {
	const (
		issuerPath   = "gno.land/r/demo/reg_auth_issuer"
		consumerPath = "gno.land/r/demo/reg_auth_consumer"
	)
	consumer := chain.PackageAddress(consumerPath)

	testing.SetRealm(testing.NewCodeRealm(issuerPath))
	_, ledger := grc721.NewToken("Authority", "AUTH", 0, cur)
	urequire.NoError(t, ledger.Mint(alice, "1"))
	urequire.NoError(t, ledger.Mint(alice, "2"))
	key := Register(cross(cur), collection.NewCollection(ledger), "")
	urequire.NoError(t, ledger.ImpersonateTeller(alice).Approve(0, cur, consumer, "1"))

	testing.SetRealm(testing.NewCodeRealm(consumerPath))
	teller := GetToken(key).RealmTeller(0, cur)
	urequire.NoError(t, teller.TransferFrom(0, cur, alice, artist, "1"))
	owner, err := MustGet(key).Token().OwnerOf("1")
	urequire.NoError(t, err)
	uassert.Equal(t, artist, owner)

	uassert.ErrorContains(t, teller.TransferFrom(0, cur, alice, artist, "2"),
		"caller is not token owner or approved")
	owner, err = MustGet(key).Token().OwnerOf("2")
	urequire.NoError(t, err)
	uassert.Equal(t, alice, owner)
}

func TestRegisteredCollectionStaysLive(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/reg_live"))
	nft, ledger := collection.New("Live", "LIVE", 0, cur)
	key := Register(cross(cur), nft.Collection(), "")
	urequire.Equal(t, 0, len(MustGet(key).Kinds()), "no extension at registration")

	royalty.NewRoyalty(ledger, 1000)

	urequire.Equal(t, 1, len(MustGet(key).Kinds()), "a later extension reaches the registry")
	_, ok := GetExtension(key, royalty.Kind)
	urequire.True(t, ok, "and the view is readable through the registry")
}

func TestUnregister(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/reg_drop"))
	_, ledger := grc721.NewToken("Drop", "DROP", 0, cur)
	key := Register(cross(cur), collection.NewCollection(ledger), "")

	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/reg_thief"))
	urequire.AbortsContains(t, cur, "collection not registered", func() {
		Unregister(cross(cur), "DROP")
	})

	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/reg_drop"))
	urequire.Equal(t, key, Unregister(cross(cur), "DROP"), "Unregister returns the dropped key")
	urequire.True(t, Get(key) == nil, "the entry is gone")

	urequire.AbortsContains(t, cur, "collection not registered", func() {
		Unregister(cross(cur), "DROP")
	})

	urequire.Equal(t, key, Register(cross(cur), collection.NewCollection(ledger), ""),
		"the freed key can be taken again")
}

func TestUnregisterFromUserRealm(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/reg_user_drop"))
	_, ledger := grc721.NewToken("User Drop", "UDROP", 0, cur)
	key := Register(cross(cur), collection.NewCollection(ledger), "")

	testing.SetRealm(testing.NewUserRealm(testutils.TestAddress("non-registrant")))
	urequire.AbortsContains(t, cur, "collection not registered", func() {
		Unregister(cross(cur), "UDROP")
	})
	urequire.True(t, Get(key) != nil, "a user caller cannot remove another realm's entry")
}

func TestRenderIndexPaginates(cur realm, t *testing.T) {
	const infoPrefix = "/r/demo/defi/grc721reg:"

	for i := 0; i < pageSize+5; i++ {
		testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/zpage" + strconv.Itoa(i)))
		_, ledger := grc721.NewToken("Page", "PAGE", 0, cur)
		Register(cross(cur), collection.NewCollection(ledger), "")
	}

	first := Render("")
	urequire.Equal(t, pageSize, strings.Count(first, infoPrefix),
		"the home page renders one page, not the whole tree")
	urequire.True(t, strings.Contains(first, "?page=2"), "the home page links the next one")

	second := Render("?page=2")
	urequire.True(t, strings.Count(second, infoPrefix) > 0, "page 2 renders the remainder")
	urequire.NotEqual(t, first, second, "page 2 is not page 1")

	// A visitor cannot widen a page, so ?size= changes nothing.
	urequire.Equal(t, first, Render("?size=1000"), "the page size is not the visitor's to set")
	urequire.Equal(t, first, Render("?page=999"), "an unavailable page falls back to page 1")
}
