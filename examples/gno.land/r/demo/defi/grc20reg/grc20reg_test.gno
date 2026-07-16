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
	Register(cross(cur), token, "mySlug")
	key := realmAddr + ".TST"
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

func TestRegistryWrappers(cur realm, t *testing.T) {
	const (
		tokenPath   = "gno.land/r/demo/token"
		realmCaller = "gno.land/r/demo/vault"
	)
	alice := testutils.TestAddress("alice")
	bob := testutils.TestAddress("bob")
	carol := testutils.TestAddress("carol")
	realmAddr := chain.PackageAddress(realmCaller)

	testing.SetRealm(testing.NewCodeRealm(tokenPath))
	token, ledger := grc20.NewToken(0, cur, "TestToken", "TST", 4)
	// #5907 keys the registry by token.ID() (pkgpath.SYMBOL), not the bare
	// pkgpath, so look up the wrappers by the token's canonical id.
	tokenKey := token.ID()
	urequire.NoError(t, ledger.Mint(alice, 1_000))
	urequire.NoError(t, ledger.Mint(realmAddr, 200))
	Register(cross(cur), token, "")
	uassert.Equal(t, "TestToken", GetName(tokenKey))
	uassert.Equal(t, "TST", GetSymbol(tokenKey))
	uassert.Equal(t, 4, GetDecimals(tokenKey))
	uassert.Equal(t, int64(1_200), TotalSupply(tokenKey))
	uassert.Equal(t, int64(1_000), BalanceOf(tokenKey, alice))
	uassert.Equal(t, int64(0), Allowance(tokenKey, alice, carol))

	testing.SetRealm(testing.NewUserRealm(alice))
	Transfer(cross(cur), tokenKey, bob, 100)
	Approve(cross(cur), tokenKey, carol, 300)

	testing.SetRealm(testing.NewUserRealm(carol))
	TransferFrom(cross(cur), tokenKey, alice, bob, 200)
	uassert.Equal(t, int64(700), BalanceOf(tokenKey, alice))
	uassert.Equal(t, int64(300), BalanceOf(tokenKey, bob))
	uassert.Equal(t, int64(100), Allowance(tokenKey, alice, carol))

	testing.SetOriginCaller(alice)
	testing.SetRealm(testing.NewCodeRealm(realmCaller))
	Transfer(cross(cur), tokenKey, bob, 50)
	uassert.Equal(t, int64(150), BalanceOf(tokenKey, realmAddr))
	uassert.Equal(t, int64(700), BalanceOf(tokenKey, alice))
	uassert.Equal(t, int64(350), BalanceOf(tokenKey, bob))

	testing.SetRealm(testing.NewUserRealm(alice))
	uassert.AbortsWithMessage(t, cur, "insufficient balance", func() {
		Transfer(cross(cur), tokenKey, bob, 701)
	})
	uassert.Equal(t, int64(700), BalanceOf(tokenKey, alice))
	uassert.Equal(t, int64(350), BalanceOf(tokenKey, bob))
}

func TestRegisterRejectsOverwrite(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/grc20reg_overwrite"))
	token, ledger := grc20.NewToken(0, cur, "Bar", "BAR", 4)
	ledger.Mint(cur.Address(), 11)
	Register(cross(cur), token, "")

	urequire.Equal(t, "BAR", Get(token.ID()).GetSymbol())
	urequire.Equal(t, int64(11), Get(token.ID()).BalanceOf(cur.Address()))

	replacement, _ := grc20.NewToken(0, cur, "Replacement", "BAR", 6)
	urequire.AbortsContains(t, cur, "token already registered", func() {
		Register(cross(cur), replacement, "")
	})
}

func TestRegisterRejectsAliasedTokenPaths(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/grc20reg_alias"))
	token, _ := grc20.NewToken(0, cur, "Aliased Token", "ALIAS", 4)
	Register(cross(cur), token, "first")

	urequire.AbortsContains(t, cur, "token already registered", func() {
		Register(cross(cur), token, "second")
	})
}

func TestRegisterRejectsTokenFromDifferentRealm(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/grc20reg_id_source"))
	token, _ := grc20.NewToken(0, cur, "Mismatch Token", "MISMATCH", 4)

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
