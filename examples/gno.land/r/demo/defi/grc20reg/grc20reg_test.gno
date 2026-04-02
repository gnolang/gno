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

func TestRegisterSlugValidation(t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/bar"))
	token, _ := grc20.NewToken("SlugTest", "SLG", 6)

	// Valid slugs
	valid := []string{"mytoken", "my-token", "my_token", "Token123", "a", "A-B_c"}
	for _, slug := range valid {
		Register(cross, token, slug)
	}

	// Invalid slugs should panic
	invalid := []string{
		`) [Claim](https://evil.com`,
		"has space",
		"has.dot",
		"has/slash",
		"[brackets]",
		"(parens)",
		"evil:colon",
	}
	for _, slug := range invalid {
		func() {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("Register with slug %q should have panicked", slug)
				}
			}()
			Register(cross, token, slug)
		}()
	}
}
