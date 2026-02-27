package grc20reg

import (
	"chain/runtime"
	"strings"
	"testing"

	"gno.land/p/demo/tokens/grc20"
	"gno.land/p/nt/ufmt"
	"gno.land/p/nt/urequire"
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

	got := Render("")
	urequire.True(t, strings.Contains(got, `- **TestToken** - [gno.land/r/demo/foo](/r/demo/foo) - [info](/r/demo/defi/grc20reg:gno.land/r/demo/foo)`))
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
	urequire.True(t, strings.Contains(got, `- **TestToken** - [gno.land/r/demo/foo](/r/demo/foo) - [info](/r/demo/defi/grc20reg:gno.land/r/demo/foo)`))
	urequire.True(t, strings.Contains(got, `- **TestToken** - [gno.land/r/demo/foo](/r/demo/foo).mySlug - [info](/r/demo/defi/grc20reg:gno.land/r/demo/foo.mySlug)`))

	expected := `# TestToken
- symbol: **TST**
- realm: [gno.land/r/demo/foo](/r/demo/foo).mySlug
- decimals: 4
- total supply: 1234567
`
	got = Render("gno.land/r/demo/foo.mySlug")
	urequire.Equal(t, expected, got)
}

func TestPagination(t *testing.T) {
	// Register 11 tokens to test pagination (default page size is 10)
	for i := 0; i < 11; i++ {
		realmPath := ufmt.Sprintf("gno.land/r/demo/token%02d", i)
		testing.SetRealm(testing.NewCodeRealm(realmPath))
		tokenName := ufmt.Sprintf("Token%02d", i)
		tokenSymbol := ufmt.Sprintf("TK%02d", i)
		token, _ := grc20.NewToken(tokenName, tokenSymbol, 6)
		Register(cross, token, "")
	}

	// Test first page (default)
	page1 := Render("")
	urequire.True(t, strings.Contains(page1, "Page 1 of 2"), "expected page 1 of 2")

	// Test second page
	page2 := Render("?page=2")
	urequire.True(t, strings.Contains(page2, "Page 2 of 2"), "expected page 2 of 2")
}
