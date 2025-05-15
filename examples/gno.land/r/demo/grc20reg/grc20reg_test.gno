package grc20reg

import (
	"std"
	"strings"
	"testing"

	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/urequire"
)

func TestRegistry(t *testing.T) {
	testing.SetRealm(std.NewCodeRealm("gno.land/r/demo/foo"))
	realmAddr := std.CurrentRealm().PkgPath()
	token, ledger := grc20.NewToken("TestToken", "TST", 4)
	ledger.Mint(std.CurrentRealm().Address(), 1234567)
	// register
	cross(Register)(token, "")
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
	cross(Register)(token, "mySlug")
	regToken = Get(realmAddr + ".mySlug")
	urequire.True(t, regToken != nil, "expected to find a token") // fixme: use urequire.NotNil
	urequire.Equal(t, regToken.GetSymbol(), "TST")

	// override
	cross(Register)(token, "")
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
