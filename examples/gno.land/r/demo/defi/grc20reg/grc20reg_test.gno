package grc20reg

import (
	"strings"
	"testing"

	"gno.land/p/demo/tokens/grc20"
	"gno.land/p/nt/urequire/v0"
)

func TestRegistry(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/foo"))
	token, ledger := grc20.NewToken("TestToken", "TST", 4, 0, cur)
	ledger.Mint(cur.Address(), 1234567)
	// register
	key := Register(cross(cur), token, "mySlug")
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

	// The registry keys by rlmPath.symbol, so a second token with the same
	// symbol in the same realm is rejected even though its Token.ID() differs
	// (distinct trailing sequence id). See TestRegisterRejectsOverwrite.
	second, _ := grc20.NewToken("Second", "TST", 4, 1, cur)
	urequire.NotEqual(t, token.ID(), second.ID()) // ids are decoupled from symbol
	urequire.AbortsContains(t, cur, "token already registered", func() {
		Register(cross(cur), second, "")
	})
}

func TestRegisterRejectsOverwrite(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/grc20reg_overwrite"))
	token, ledger := grc20.NewToken("Bar", "BAR", 4, 0, cur)
	ledger.Mint(cur.Address(), 11)
	key := Register(cross(cur), token, "")

	urequire.Equal(t, "BAR", Get(key).GetSymbol())
	urequire.Equal(t, int64(11), Get(key).BalanceOf(cur.Address()))

	replacement, _ := grc20.NewToken("Replacement", "BAR", 6, 0, cur)
	urequire.AbortsContains(t, cur, "token already registered", func() {
		Register(cross(cur), replacement, "")
	})
}

func TestRegisterRejectsAliasedTokenPaths(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/grc20reg_alias"))
	token, _ := grc20.NewToken("Aliased Token", "ALIAS", 4, 0, cur)
	Register(cross(cur), token, "first")

	urequire.AbortsContains(t, cur, "token already registered", func() {
		Register(cross(cur), token, "second")
	})
}

func TestRegisterRejectsTokenFromDifferentRealm(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/grc20reg_id_source"))
	token, _ := grc20.NewToken("Mismatch Token", "MISMATCH", 4, 0, cur)

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
