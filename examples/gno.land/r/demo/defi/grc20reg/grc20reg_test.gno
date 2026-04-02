package grc20reg

import (
	"chain/runtime"
	"strings"
	"testing"

	"gno.land/p/demo/tokens/grc20"
	"gno.land/p/nt/urequire/v0"
)

func TestRegistry(t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/foo"))
	realmAddr := runtime.CurrentRealm().PkgPath()
	token, ledger := grc20.NewToken("TestToken", "TST", 4)
	ledger.Mint(runtime.CurrentRealm().Address(), 1234567)
	// register
	Register(cross, token, "")
	regToken := Get(realmAddr)
	urequire.True(t, regToken != nil, "expected to find a token") // fixme: use urequire.NotNil
	urequire.Equal(t, regToken.GetSymbol(), "TST")

	expected := `- **TestToken** - [gno.land/r/demo/foo](/r/demo/foo) - [info](/r/demo/grc20reg:gno.land/r/demo/foo)
`
	got := Render("")
	urequire.True(t, strings.Contains(got, expected))
	// 404
	invalidToken := Get("0xdeadbeef")
	urequire.True(t, invalidToken == nil)

	// register with a slug
	Register(cross, token, "mySlug")
	regToken = Get(realmAddr + ".mySlug")
	urequire.True(t, regToken != nil, "expected to find a token") // fixme: use urequire.NotNil
	urequire.Equal(t, regToken.GetSymbol(), "TST")

	// override
	Register(cross, token, "")
	regToken = Get(realmAddr + "")
	urequire.True(t, regToken != nil, "expected to find a token") // fixme: use urequire.NotNil
	urequire.Equal(t, regToken.GetSymbol(), "TST")

	got = Render("")
	urequire.True(t, strings.Contains(got, `- **TestToken** - [gno.land/r/demo/foo](/r/demo/foo) - [info](/r/demo/grc20reg:gno.land/r/demo/foo)`))
	urequire.True(t, strings.Contains(got, `- **TestToken** - [gno.land/r/demo/foo](/r/demo/foo).mySlug - [info](/r/demo/grc20reg:gno.land/r/demo/foo.mySlug)`))

	expected = `# TestToken
- symbol: **TST**
- realm: [gno.land/r/demo/foo](/r/demo/foo).mySlug
- decimals: 4
- total supply: 1234567
`
	got = Render("gno.land/r/demo/foo.mySlug")
	urequire.Equal(t, expected, got)
}

func TestValidateSlug(t *testing.T) {
	// Valid slugs — should not panic
	valid := []string{"mytoken", "my-token", "my_token", "Token123", "a", "A-B_c"}
	for _, slug := range valid {
		validateSlug(slug) // no panic = pass
	}
}

func TestValidateSlugPanicsOnSpace(t *testing.T) {
	defer func() { recover() }()
	validateSlug("has space")
	t.Errorf("should have panicked")
}

func TestValidateSlugPanicsOnDot(t *testing.T) {
	defer func() { recover() }()
	validateSlug("has.dot")
	t.Errorf("should have panicked")
}

func TestValidateSlugPanicsOnSlash(t *testing.T) {
	defer func() { recover() }()
	validateSlug("has/slash")
	t.Errorf("should have panicked")
}

func TestValidateSlugPanicsOnBrackets(t *testing.T) {
	defer func() { recover() }()
	validateSlug("[brackets]")
	t.Errorf("should have panicked")
}

func TestValidateSlugPanicsOnParens(t *testing.T) {
	defer func() { recover() }()
	validateSlug("(parens)")
	t.Errorf("should have panicked")
}

func TestValidateSlugPanicsOnInjection(t *testing.T) {
	defer func() { recover() }()
	validateSlug(`) [Claim](https://evil.com`)
	t.Errorf("should have panicked")
}
