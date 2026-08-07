package grc20reg

import (
	"strings"
	"testing"

	"gno.land/p/demo/tokens/grc20"
	"gno.land/p/nt/seqid/v0"
	"gno.land/p/nt/urequire/v0"
)

func TestRegistry(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/foo"))
	token, ledger := grc20.NewToken("TestToken", "TST", 4, IdentifierGenerator(cross(cur)), cur)
	ledger.Mint(cur.Address(), 1234567)
	// The key is the (realm, symbol) pair; Token.ID() additionally carries the
	// generator id, so the two are distinct values.
	key := Register(cross(cur), token, "mySlug")
	urequire.Equal(t, "gno.land/r/demo/foo.TST", key)
	urequire.NotEqual(t, token.ID(), key)
	regToken := Get(key)
	urequire.True(t, regToken != nil, "expected to find a token") // fixme: use urequire.NotNil
	urequire.Equal(t, regToken.GetSymbol(), "TST")

	expected := `- **TestToken** - [gno.land/r/demo/foo](/r/demo/foo).TST - [info](/r/demo/grc20reg:gno.land/r/demo/foo.TST)
`
	got := Render("")
	urequire.True(t, strings.Contains(got, expected))
	// 404
	invalidToken := Get("0xdeadbeef")
	urequire.True(t, invalidToken == nil)

	got = Render("")
	urequire.True(t, strings.Contains(got, expected))

	expected = `# TestToken
- symbol: **TST**
- realm: [gno.land/r/demo/foo](/r/demo/foo).TST
- decimals: 4
- total supply: 1234567
`
	got = Render(key)
	urequire.Equal(t, expected, got)
}

func TestRegisterRejectsDuplicateSymbol(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/grc20reg_dup"))
	first, _ := grc20.NewToken("First", "DUP", 4, IdentifierGenerator(cross(cur)), cur)
	second, _ := grc20.NewToken("Second", "DUP", 4, IdentifierGenerator(cross(cur)), cur)

	// The two tokens have distinct ids — each drew its own id — but they
	// share a (realm, symbol) key, so only the first can be registered.
	urequire.NotEqual(t, first.ID(), second.ID())
	Register(cross(cur), first, "")
	urequire.AbortsContains(t, cur, "token already registered", func() {
		Register(cross(cur), second, "")
	})
}

func TestRegisterRejectsSelfIssuedID(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/grc20reg_self_issued"))
	// A realm can always create its own generator, but PackagePath() is bound
	// under IsCurrent, so it cannot claim to be the registry's.
	var ownSeq seqid.ID
	own := grc20.NewIDGenerator(0, cur, func() string { return ownSeq.Next().String() })
	token, _ := grc20.NewToken("Foreign", "FGN", 4, own, cur)

	urequire.AbortsContains(t, cur, "not issued by the registry generator", func() {
		Register(cross(cur), token, "")
	})
}

func TestCopyingTheGeneratorCannotReplayAnID(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/grc20reg_replay"))

	// The #6026 reproduction: the generator holds only a closure over the
	// registry's counter, so copies of it advance that same counter.
	gen := IdentifierGenerator(cross(cur))
	a, b := *gen, *gen

	canonical, _ := grc20.NewToken("Canonical", "DUP", 6, &a, cur)
	shadow, _ := grc20.NewToken("Shadow", "DUP", 6, &b, cur)
	urequire.NotEqual(t, canonical.ID(), shadow.ID())

	// Same for a freshly created one.
	other, _ := grc20.NewToken("Other", "DUP", 6, IdentifierGenerator(cross(cur)), cur)
	urequire.NotEqual(t, canonical.ID(), other.ID())
	urequire.NotEqual(t, shadow.ID(), other.ID())
}

func TestSelfIssuedCannotReplayARegistryID(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/grc20reg_replay_self"))

	// The registry's counter is public chain state, so a realm can read the id
	// off a registered token and have its own generator return that id verbatim.
	// The issuer path is part of Token.ID(), so the two still differ.
	canonical, _ := grc20.NewToken("Canonical", "RPL", 6, IdentifierGenerator(cross(cur)), cur)
	id := canonical.ID()[strings.LastIndex(canonical.ID(), ":")+1:]

	own := grc20.NewIDGenerator(0, cur, func() string { return id })
	shadow, _ := grc20.NewToken("Shadow", "RPL", 6, own, cur)
	urequire.NotEqual(t, canonical.ID(), shadow.ID())

	// Nor can it forge the registry's form: ':' is outside the id charset.
	urequire.PanicsContains(t, cur, "invalid id", func() {
		forged := grc20.NewIDGenerator(0, cur, func() string { return cur.PkgPath() + ":" + id })
		grc20.NewToken("Forged", "RPL", 6, forged, cur)
	})
}

func TestRegisterRejectsTokenFromDifferentRealm(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/grc20reg_id_source"))
	token, _ := grc20.NewToken("Mismatch Token", "MISMATCH", 4, IdentifierGenerator(cross(cur)), cur)

	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/grc20reg_id_target"))
	urequire.AbortsContains(t, cur, "token must be registered from its own realm", func() {
		Register(cross(cur), token, "")
	})
}

func TestValidateSlug(cur realm, t *testing.T) {
	// Valid slugs — should not panic
	valid := []string{"mytoken", "my-token", "my_token", "Token123", "a", "A-B_c", strings.Repeat("a", maxSlugLen)}
	for _, slug := range valid {
		validateSlug(slug) // no panic = pass
	}
}

func TestValidateSlugPanicsOnTooLong(cur realm, t *testing.T) {
	defer func() { recover() }()
	validateSlug(strings.Repeat("a", maxSlugLen+1))
	t.Errorf("should have panicked")
}

func TestValidateSlugPanicsOnSpace(cur realm, t *testing.T) {
	defer func() { recover() }()
	validateSlug("has space")
	t.Errorf("should have panicked")
}

func TestValidateSlugPanicsOnDot(cur realm, t *testing.T) {
	defer func() { recover() }()
	validateSlug("has.dot")
	t.Errorf("should have panicked")
}

func TestValidateSlugPanicsOnSlash(cur realm, t *testing.T) {
	defer func() { recover() }()
	validateSlug("has/slash")
	t.Errorf("should have panicked")
}

func TestValidateSlugPanicsOnBrackets(cur realm, t *testing.T) {
	defer func() { recover() }()
	validateSlug("[brackets]")
	t.Errorf("should have panicked")
}

func TestValidateSlugPanicsOnParens(cur realm, t *testing.T) {
	defer func() { recover() }()
	validateSlug("(parens)")
	t.Errorf("should have panicked")
}

func TestValidateSlugPanicsOnInjection(cur realm, t *testing.T) {
	defer func() { recover() }()
	validateSlug(`) [Claim](https://evil.com`)
	t.Errorf("should have panicked")
}

func TestRegisterRejectsNilToken(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/grc20reg_nil"))
	urequire.AbortsContains(t, cur, "nil token", func() {
		Register(cross(cur), nil, "")
	})
}
