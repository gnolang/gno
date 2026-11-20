package grc20reg

import (
	"chain"
	"strings"
	"testing"

	"gno.land/p/nt/grc20/v0"
	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/ufmt/v0"
	"gno.land/p/nt/urequire/v0"
)

// tokenFuncs are the token realm's crossing functions. This test package's
// realm plays the token realm for its fixtures, so they are declared here.
func tokenFuncs(ledger *grc20.PrivateLedger) (
	func(cur realm, to address, amount int64),
	func(cur realm, spender address, amount int64),
	func(cur realm, from, to address, amount int64),
) {
	transfer := func(cur realm, to address, amount int64) {
		checkErr(ledger.CallerTeller().Transfer(0, cur, to, amount))
	}
	approve := func(cur realm, spender address, amount int64) {
		checkErr(ledger.CallerTeller().Approve(0, cur, spender, amount))
	}
	transferFrom := func(cur realm, from, to address, amount int64) {
		checkErr(ledger.CallerTeller().TransferFrom(0, cur, from, to, amount))
	}
	return transfer, approve, transferFrom
}

func TestRegistry(cur realm, t *testing.T) {
	token, ledger := grc20.NewToken("TestToken", "TST", 4, 0, cur)
	ledger.Mint(cur.Address(), 1234567)
	transfer, approve, transferFrom := tokenFuncs(ledger)
	key := Register(cross(cur), token, "mySlug", transfer, approve, transferFrom)
	regToken := Get(key)
	urequire.True(t, regToken != nil, "expected to find a token")
	urequire.Equal(t, regToken.GetSymbol(), "TST")

	expected := `- **TestToken** - [gno.land/r/nt/grc20reg/v0](/r/nt/grc20reg/v0).TST - [info](/r/nt/grc20reg/v0:gno.land/r/nt/grc20reg/v0.TST)
`
	got := Render("")
	urequire.True(t, strings.Contains(got, expected))
	invalidToken := Get("0xdeadbeef")
	urequire.True(t, invalidToken == nil)

	expected = `# TestToken
- symbol: **TST**
- realm: [gno.land/r/nt/grc20reg/v0](/r/nt/grc20reg/v0).TST
- decimals: 4
- total supply: 1234567
`
	got = Render(key)
	urequire.Equal(t, expected, got)

	second, secondLedger := grc20.NewToken("Second", "TST", 4, 1, cur)
	st, sa, stf := tokenFuncs(secondLedger)
	urequire.NotEqual(t, token.ID(), second.ID())
	urequire.AbortsContains(t, cur, "token already registered", func() {
		Register(cross(cur), second, "", st, sa, stf)
	})
}

// TestRegistryLookupIsReadOnly pins that Get/MustGet hand out a token for
// reading only: a foreign realm can observe balances, but the write path is the
// token realm's crossing functions, reached through the registry wrappers.
func TestRegistryLookupIsReadOnly(cur realm, t *testing.T) {
	const consumerPath = "gno.land/r/demo/grc20reg_reader"
	alice := testutils.TestAddress("alice")

	token, ledger := grc20.NewToken("ReadOnly", "RONLY", 4, 0, cur)
	urequire.NoError(t, ledger.Mint(alice, 1_000))
	transfer, approve, transferFrom := tokenFuncs(ledger)
	tokenKey := Register(cross(cur), token, "", transfer, approve, transferFrom)
	uassert.Equal(t, "ReadOnly", MustGet(tokenKey).GetName())

	testing.SetRealm(testing.NewCodeRealm(consumerPath))
	uassert.Equal(t, int64(1_000), MustGet(tokenKey).BalanceOf(alice))
	uassert.Equal(t, int64(0), MustGet(tokenKey).Allowance(alice, chain.PackageAddress(consumerPath)))
}

func TestRegisterRejectsOverwrite(cur realm, t *testing.T) {
	token, ledger := grc20.NewToken("Bar", "BAR", 4, 0, cur)
	ledger.Mint(cur.Address(), 11)
	transfer, approve, transferFrom := tokenFuncs(ledger)
	key := Register(cross(cur), token, "", transfer, approve, transferFrom)

	urequire.Equal(t, "BAR", Get(key).GetSymbol())
	urequire.Equal(t, int64(11), Get(key).BalanceOf(cur.Address()))

	replacement, replacementLedger := grc20.NewToken("Replacement", "BAR", 6, 0, cur)
	rt, ra, rtf := tokenFuncs(replacementLedger)
	urequire.AbortsContains(t, cur, "token already registered", func() {
		Register(cross(cur), replacement, "", rt, ra, rtf)
	})
}

func TestRegisterRejectsAliasedTokenPaths(cur realm, t *testing.T) {
	token, ledger := grc20.NewToken("Aliased Token", "ALIAS", 4, 0, cur)
	transfer, approve, transferFrom := tokenFuncs(ledger)
	Register(cross(cur), token, "first", transfer, approve, transferFrom)

	urequire.AbortsContains(t, cur, "token already registered", func() {
		Register(cross(cur), token, "second", transfer, approve, transferFrom)
	})
}

func TestRegisterRejectsTokenFromDifferentRealm(cur realm, t *testing.T) {
	token, ledger := grc20.NewToken("Mismatch Token", "MISMATCH", 4, 0, cur)
	transfer, approve, transferFrom := tokenFuncs(ledger)

	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/grc20reg_id_target"))
	urequire.AbortsContains(t, cur, "token must be registered from its own realm", func() {
		Register(cross(cur), token, "", transfer, approve, transferFrom)
	})
}

func TestPagination(cur realm, t *testing.T) {
	// One token per realm+symbol, so span the page size across distinct realms.
	for i := 0; i < 11; i++ {
		testing.SetRealm(testing.NewCodeRealm(ufmt.Sprintf("gno.land/r/demo/token%02d", i)))
		token, _ := grc20.NewToken(ufmt.Sprintf("Token%02d", i), ufmt.Sprintf("TK%02d", i), 6, 0, cur)
		Register(cross(cur), token, "")
	}

	// Other tests in this file share the registry, so derive the page count
	// from its current size instead of hardcoding it.
	totalPages := (registry.Size() + 9) / 10
	urequire.True(t, totalPages > 1, "expected more than one page")

	page1 := Render("")
	urequire.True(t, strings.Contains(page1, ufmt.Sprintf("Page 1 of %d", totalPages)))

	page2 := Render("?page=2")
	urequire.True(t, strings.Contains(page2, ufmt.Sprintf("Page 2 of %d", totalPages)))

	// The picker must carry a custom page size across page links.
	sized := Render("?size=3")
	urequire.True(t, strings.Contains(sized, "(?page=2&size=3)"), "picker links must keep size")
}

func TestValidateSlug(cur realm, t *testing.T) {
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
	urequire.AbortsContains(t, cur, "nil token", func() {
		Register(cross(cur), nil, "", nil, nil, nil)
	})
}

func TestRegisterRejectsNilCrossingFunction(cur realm, t *testing.T) {
	token, _ := grc20.NewToken("NilFn", "NILFN", 4, 0, cur)
	urequire.AbortsContains(t, cur, "nil crossing function", func() {
		Register(cross(cur), token, "", nil, nil, nil)
	})
}

// TestWrappersBindActorToCallingRealm: the wrappers are non-crossing, so the
// caller forwards its own live cur and the token realm's crossing function acts
// for the caller, not for the registry.
func TestWrappersBindActorToCallingRealm(cur realm, t *testing.T) {
	const consumerPath = "gno.land/r/demo/grc20reg_wrapper_consumer"
	bob := testutils.TestAddress("bob")
	consumer := chain.PackageAddress(consumerPath)
	registry := chain.PackageAddress("gno.land/r/nt/grc20reg/v0")

	token, ledger := grc20.NewToken("WrapperActor", "WRPA", 4, 0, cur)
	urequire.NoError(t, ledger.Mint(consumer, 1_000))
	urequire.NoError(t, ledger.Mint(registry, 1_000))
	transfer, approve, transferFrom := tokenFuncs(ledger)
	tokenKey := Register(cross(cur), token, "", transfer, approve, transferFrom)

	testing.SetRealm(testing.NewCodeRealm(consumerPath))
	Transfer(0, cur, tokenKey, bob, 100)

	uassert.Equal(t, int64(900), MustGet(tokenKey).BalanceOf(consumer))
	uassert.Equal(t, int64(1_000), MustGet(tokenKey).BalanceOf(registry))
	uassert.Equal(t, int64(100), MustGet(tokenKey).BalanceOf(bob))
}

func TestWrapperActorIgnoresSigningUser(cur realm, t *testing.T) {
	const consumerPath = "gno.land/r/demo/grc20reg_wrapper_origin_consumer"
	alice := testutils.TestAddress("alice")
	bob := testutils.TestAddress("bob")
	consumer := chain.PackageAddress(consumerPath)

	token, ledger := grc20.NewToken("WrapperOrigin", "WRPO", 4, 0, cur)
	urequire.NoError(t, ledger.Mint(alice, 1_000))
	urequire.NoError(t, ledger.Mint(consumer, 500))
	transfer, approve, transferFrom := tokenFuncs(ledger)
	tokenKey := Register(cross(cur), token, "", transfer, approve, transferFrom)

	testing.SetOriginCaller(alice)
	testing.SetRealm(testing.NewCodeRealm(consumerPath))
	Transfer(0, cur, tokenKey, bob, 100)

	uassert.Equal(t, int64(1_000), MustGet(tokenKey).BalanceOf(alice))
	uassert.Equal(t, int64(400), MustGet(tokenKey).BalanceOf(consumer))
}

func TestTransferFromSpendsAllowanceGrantedToCallingRealm(cur realm, t *testing.T) {
	const consumerPath = "gno.land/r/demo/grc20reg_wrapper_allowance_consumer"
	alice := testutils.TestAddress("alice")
	bob := testutils.TestAddress("bob")
	consumer := chain.PackageAddress(consumerPath)

	token, ledger := grc20.NewToken("WrapperAllowance", "WRPL", 4, 0, cur)
	urequire.NoError(t, ledger.Mint(alice, 1_000))
	transfer, approve, transferFrom := tokenFuncs(ledger)
	tokenKey := Register(cross(cur), token, "", transfer, approve, transferFrom)

	testing.SetRealm(testing.NewCodeRealm(consumerPath))
	uassert.AbortsContains(t, cur, "insufficient allowance", func() {
		TransferFrom(0, cur, tokenKey, alice, bob, 100)
	})
	uassert.Equal(t, int64(1_000), MustGet(tokenKey).BalanceOf(alice))

	urequire.NoError(t, ledger.Approve(alice, consumer, 100))

	testing.SetRealm(testing.NewCodeRealm(consumerPath))
	TransferFrom(0, cur, tokenKey, alice, bob, 100)
	uassert.Equal(t, int64(900), MustGet(tokenKey).BalanceOf(alice))
	uassert.Equal(t, int64(100), MustGet(tokenKey).BalanceOf(bob))
}

func TestGetRegistryYieldsTokens(cur realm, t *testing.T) {
	token, ledger := grc20.NewToken("RegistryView", "RVW", 4, 0, cur)
	transfer, approve, transferFrom := tokenFuncs(ledger)
	key := Register(cross(cur), token, "", transfer, approve, transferFrom)

	value := GetRegistry().Get(key)
	regToken, ok := value.(*grc20.Token)
	urequire.True(t, ok, "expected a *grc20.Token")
	urequire.Equal(t, "RVW", regToken.GetSymbol())
}
