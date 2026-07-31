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
	urequire.Equal(t, "gno.land/r/demo/foo.mySlug", key)
	regToken := Get(key)
	urequire.True(t, regToken != nil, "expected to find a token") // fixme: use urequire.NotNil
	urequire.Equal(t, regToken.GetSymbol(), "TST")

	expected := `- **TestToken** - [gno.land/r/demo/foo](/r/demo/foo).mySlug - [info](/r/demo/grc20reg:gno.land/r/demo/foo.mySlug)
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
- realm: [gno.land/r/demo/foo](/r/demo/foo).mySlug
- decimals: 4
- total supply: 1234567
`
	got = Render(key)
	urequire.Equal(t, expected, got)

	// Tokens with the same symbol can coexist under different aliases.
	second, _ := grc20.NewToken("Second", "TST", 4, 1, cur)
	urequire.NotEqual(t, token.ID(), second.ID())
	secondKey := Register(cross(cur), second, "second")
	urequire.Equal(t, "gno.land/r/demo/foo.second", secondKey)
	urequire.Equal(t, second.ID(), Get(secondKey).ID())
}

func TestRegisterRejectsOverwrite(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/grc20reg_overwrite"))
	token, ledger := grc20.NewToken("Bar", "BAR", 4, 0, cur)
	ledger.Mint(cur.Address(), 11)
	key := Register(cross(cur), token, "shared")

	urequire.Equal(t, "BAR", Get(key).GetSymbol())
	urequire.Equal(t, int64(11), Get(key).BalanceOf(cur.Address()))

	replacement, _ := grc20.NewToken("Replacement", "OTHER", 6, 1, cur)
	urequire.AbortsContains(t, cur, "token already registered", func() {
		Register(cross(cur), replacement, "shared")
	})
	urequire.Equal(t, "Bar", Get(key).GetName())
}

func TestRegisterAllowsMultipleAliases(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/grc20reg_alias"))
	token, _ := grc20.NewToken("Aliased Token", "ALIAS", 4, 0, cur)

	first := Register(cross(cur), token, "first")
	second := Register(cross(cur), token, "second")

	urequire.Equal(t, token.ID(), Get(first).ID())
	urequire.Equal(t, token.ID(), Get(second).ID())
}

func TestRegisterRejectsTokenFromDifferentRealm(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/grc20reg_id_source"))
	token, _ := grc20.NewToken("Mismatch Token", "MISMATCH", 4, 0, cur)

	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/grc20reg_id_target"))
	urequire.AbortsContains(t, cur, "token must be registered from its own realm", func() {
		Register(cross(cur), token, "mismatch")
	})
}

func TestRegisterRejectsEmptySlug(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/grc20reg_empty"))
	token, _ := grc20.NewToken("Empty Slug", "EMPTY", 4, 0, cur)

	urequire.AbortsContains(t, cur, "empty slug", func() {
		Register(cross(cur), token, "")
	})
}

func TestRegisterPanicsOnInvalidSlug(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/grc20reg_invalid"))
	token, _ := grc20.NewToken("Invalid Slug", "INVALID", 4, 0, cur)

	urequire.AbortsContains(t, cur, "slug too long", func() {
		Register(cross(cur), token, strings.Repeat("a", maxSlugLen+1))
	})
}

func TestValidateSlug(cur realm, t *testing.T) {
	valid := []string{"mytoken", "my-token", "my_token", "Token123", "a", "A-B_c", strings.Repeat("a", maxSlugLen)}
	for _, slug := range valid {
		urequire.NoError(t, validateSlug(slug))
	}

	invalid := []string{
		strings.Repeat("a", maxSlugLen+1),
		"has space",
		"has.dot",
		"has/slash",
		"[brackets]",
		"(parens)",
		`) [Claim](https://evil.com`,
	}
	for _, slug := range invalid {
		urequire.Error(t, validateSlug(slug))
	}
}

func TestRegisterRejectsNilToken(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/grc20reg_nil"))
	urequire.AbortsContains(t, cur, "nil token", func() {
		Register(cross(cur), nil, "")
	})
}
