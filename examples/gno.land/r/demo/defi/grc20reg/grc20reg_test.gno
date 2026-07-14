package grc20reg

import (
	"chain"
	"strings"
	"testing"

	"gno.land/p/demo/tokens/grc20"
	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"
)

func TestRegistry(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/foo"))
	realmAddr := cur.PkgPath()
	token, ledger := grc20.NewToken(0, cur, "TestToken", "TST", 4)
	ledger.Mint(cur.Address(), 1234567)
	// register
	Register(cross(cur), token, "")
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
	Register(cross(cur), token, "mySlug")
	regToken = Get(realmAddr + ".mySlug")
	urequire.True(t, regToken != nil, "expected to find a token") // fixme: use urequire.NotNil
	urequire.Equal(t, regToken.GetSymbol(), "TST")

	// override
	Register(cross(cur), token, "")
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

func TestRegistryWrappers(cur realm, t *testing.T) {
	const (
		tokenKey    = "gno.land/r/demo/token"
		realmCaller = "gno.land/r/demo/vault"
	)
	alice := testutils.TestAddress("alice")
	bob := testutils.TestAddress("bob")
	carol := testutils.TestAddress("carol")
	realmAddr := chain.PackageAddress(realmCaller)

	testing.SetRealm(testing.NewCodeRealm(tokenKey))
	token, ledger := grc20.NewToken(0, cur, "TestToken", "TST", 4)
	urequire.NoError(t, ledger.Mint(alice, 1_000))
	urequire.NoError(t, ledger.Mint(realmAddr, 200))
	Register(cross(cur), token, "")

	testing.SetRealm(testing.NewUserRealm(alice))
	Transfer(cross(cur), tokenKey, bob, 100)
	Approve(cross(cur), tokenKey, carol, 300)

	testing.SetRealm(testing.NewUserRealm(carol))
	TransferFrom(cross(cur), tokenKey, alice, bob, 200)
	uassert.Equal(t, int64(700), token.BalanceOf(alice))
	uassert.Equal(t, int64(300), token.BalanceOf(bob))
	uassert.Equal(t, int64(100), token.Allowance(alice, carol))

	testing.SetOriginCaller(alice)
	testing.SetRealm(testing.NewCodeRealm(realmCaller))
	Transfer(cross(cur), tokenKey, bob, 50)
	uassert.Equal(t, int64(150), token.BalanceOf(realmAddr))
	uassert.Equal(t, int64(700), token.BalanceOf(alice))
	uassert.Equal(t, int64(350), token.BalanceOf(bob))

	testing.SetRealm(testing.NewUserRealm(alice))
	uassert.AbortsWithMessage(t, cur, "insufficient balance", func() {
		Transfer(cross(cur), tokenKey, bob, 701)
	})
	uassert.Equal(t, int64(700), token.BalanceOf(alice))
	uassert.Equal(t, int64(350), token.BalanceOf(bob))
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
