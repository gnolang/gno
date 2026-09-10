package grc20reg

import (
	"chain"
	"strings"
	"testing"

	"gno.land/p/demo/tokens/grc20"
	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/ufmt/v0"
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

	expected := `- **TestToken** - [gno.land/r/demo/foo](/r/demo/foo).TST - [info](/r/demo/defi/grc20reg:gno.land/r/demo/foo.TST)
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

// TestRegistryLookupGrantsNoSpendAuthority pins the property that replaced the
// former Transfer/Approve/TransferFrom wrappers: a registry lookup yields a
// *Token, and a *Token alone carries no authority to debit anybody. The
// frame-relative teller is reachable only from the ledger, which never leaves
// the token's own realm, so this realm cannot act on behalf of its caller —
// that shape was the confused deputy.
//
// The supported route for a realm that must move a user's funds is the one
// exercised below: the user grants an allowance, and the realm spends it as
// itself through a RealmTeller, which is eagerly bound to its own address.
func TestRegistryLookupGrantsNoSpendAuthority(cur realm, t *testing.T) {
	const (
		tokenPath    = "gno.land/r/demo/token"
		consumerPath = "gno.land/r/demo/grc20reg_consumer"
	)
	alice := testutils.TestAddress("alice")
	bob := testutils.TestAddress("bob")
	consumer := chain.PackageAddress(consumerPath)

	testing.SetRealm(testing.NewCodeRealm(tokenPath))
	token, ledger := grc20.NewToken("TestToken", "TST", 4, 0, cur)
	urequire.NoError(t, ledger.Mint(alice, 1_000))
	tokenKey := Register(cross(cur), token, "")
	uassert.Equal(t, "TestToken", MustGet(tokenKey).GetName())
	uassert.Equal(t, int64(1_000), MustGet(tokenKey).BalanceOf(alice))
	uassert.Equal(t, int64(0), MustGet(tokenKey).Allowance(alice, consumer))

	// alice approves the consumer realm on the token. On chain she does this
	// through the token realm's own entry point; the ledger stands in for it.
	urequire.NoError(t, ledger.ImpersonateTeller(alice).Approve(0, cur, consumer, 300))
	uassert.Equal(t, int64(300), MustGet(tokenKey).Allowance(alice, consumer))

	// The consumer spends that allowance as ITSELF, over a token it found in
	// the registry and does not own. The actor is fixed at construction, so it
	// cannot be redirected by whoever calls in.
	testing.SetRealm(testing.NewCodeRealm(consumerPath))
	spender := MustGet(tokenKey).RealmTeller(0, cur)
	urequire.NoError(t, spender.TransferFrom(0, cur, alice, bob, 200))
	uassert.Equal(t, int64(800), MustGet(tokenKey).BalanceOf(alice))
	uassert.Equal(t, int64(200), MustGet(tokenKey).BalanceOf(bob))
	uassert.Equal(t, int64(100), MustGet(tokenKey).Allowance(alice, consumer))

	// Beyond the allowance it stops: the balance is not reachable directly.
	uassert.ErrorContains(t,
		spender.TransferFrom(0, cur, alice, bob, 101),
		"insufficient allowance")
	uassert.Equal(t, int64(800), MustGet(tokenKey).BalanceOf(alice))
	uassert.Equal(t, int64(200), MustGet(tokenKey).BalanceOf(bob))
}

// TestRegistryConsumerCannotRedirectItsActor is the negative half of the
// property above: a realm holding a registered token pays out of its own
// balance, not the signing user's. The actor is bound at construction, so the
// transaction's origin caller has no bearing on who is debited — which is what
// made the frame-relative teller in a foreign realm a phishing primitive.
func TestRegistryConsumerCannotRedirectItsActor(cur realm, t *testing.T) {
	const (
		tokenPath    = "gno.land/r/demo/token_actor_binding"
		consumerPath = "gno.land/r/demo/grc20reg_actor_consumer"
	)
	alice := testutils.TestAddress("alice")
	bob := testutils.TestAddress("bob")
	consumer := chain.PackageAddress(consumerPath)

	testing.SetRealm(testing.NewCodeRealm(tokenPath))
	token, ledger := grc20.NewToken("ActorBinding", "ACTB", 4, 0, cur)
	urequire.NoError(t, ledger.Mint(alice, 1_000))
	urequire.NoError(t, ledger.Mint(consumer, 500))
	tokenKey := Register(cross(cur), token, "")

	// alice signs the transaction, and holds a balance the consumer would love
	// to spend. The consumer pays out of its own instead.
	testing.SetOriginCaller(alice)
	testing.SetRealm(testing.NewCodeRealm(consumerPath))
	urequire.NoError(t, MustGet(tokenKey).RealmTeller(0, cur).Transfer(0, cur, bob, 100))
	uassert.Equal(t, int64(1_000), MustGet(tokenKey).BalanceOf(alice))
	uassert.Equal(t, int64(400), MustGet(tokenKey).BalanceOf(consumer))
	uassert.Equal(t, int64(100), MustGet(tokenKey).BalanceOf(bob))
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

// TestWrappersBindActorToCallingRealm pins the property that makes the write
// wrappers safe: they are non-crossing, so `rlm` reaches RealmTeller as the
// caller's own live token and the debit lands on the caller. A crossing wrapper
// would mint a fresh `cur` for grc20reg and spend the registry's balance
// instead, which is why the convention is not a stylistic choice.
//
// Both the consumer and the registry hold a balance, so whichever one is
// debited is observable rather than inferred.
func TestWrappersBindActorToCallingRealm(cur realm, t *testing.T) {
	const (
		tokenPath    = "gno.land/r/demo/token_wrapper_actor"
		consumerPath = "gno.land/r/demo/grc20reg_wrapper_consumer"
	)
	bob := testutils.TestAddress("bob")
	consumer := chain.PackageAddress(consumerPath)
	registry := chain.PackageAddress("gno.land/r/demo/defi/grc20reg")

	testing.SetRealm(testing.NewCodeRealm(tokenPath))
	token, ledger := grc20.NewToken("WrapperActor", "WRPA", 4, 0, cur)
	urequire.NoError(t, ledger.Mint(consumer, 1_000))
	urequire.NoError(t, ledger.Mint(registry, 1_000))
	tokenKey := Register(cross(cur), token, "")

	testing.SetRealm(testing.NewCodeRealm(consumerPath))
	Transfer(0, cur, tokenKey, bob, 100)

	uassert.Equal(t, int64(900), MustGet(tokenKey).BalanceOf(consumer))
	uassert.Equal(t, int64(1_000), MustGet(tokenKey).BalanceOf(registry))
	uassert.Equal(t, int64(100), MustGet(tokenKey).BalanceOf(bob))
}

// TestWrapperActorIgnoresSigningUser is the negative half: the signing user has
// a balance the calling realm would like to spend, and does not lose it. The
// actor comes from rlm.Address(), so the transaction's origin has no say in who
// pays — a hub cannot be induced to debit its caller's caller.
func TestWrapperActorIgnoresSigningUser(cur realm, t *testing.T) {
	const (
		tokenPath    = "gno.land/r/demo/token_wrapper_origin"
		consumerPath = "gno.land/r/demo/grc20reg_wrapper_origin_consumer"
	)
	alice := testutils.TestAddress("alice")
	bob := testutils.TestAddress("bob")
	consumer := chain.PackageAddress(consumerPath)

	testing.SetRealm(testing.NewCodeRealm(tokenPath))
	token, ledger := grc20.NewToken("WrapperOrigin", "WRPO", 4, 0, cur)
	urequire.NoError(t, ledger.Mint(alice, 1_000))
	urequire.NoError(t, ledger.Mint(consumer, 500))
	tokenKey := Register(cross(cur), token, "")

	testing.SetOriginCaller(alice)
	testing.SetRealm(testing.NewCodeRealm(consumerPath))
	Transfer(0, cur, tokenKey, bob, 100)

	uassert.Equal(t, int64(1_000), MustGet(tokenKey).BalanceOf(alice))
	uassert.Equal(t, int64(400), MustGet(tokenKey).BalanceOf(consumer))
}

// TestTransferFromSpendsAllowanceGrantedToCallingRealm records the semantic
// shift that comes with binding the actor to the caller: the owner must have
// approved the CALLING REALM, not the signing user. An allowance granted to
// anyone else does not authorize the wrapper.
func TestTransferFromSpendsAllowanceGrantedToCallingRealm(cur realm, t *testing.T) {
	const (
		tokenPath    = "gno.land/r/demo/token_wrapper_allowance"
		consumerPath = "gno.land/r/demo/grc20reg_wrapper_allowance_consumer"
	)
	alice := testutils.TestAddress("alice")
	bob := testutils.TestAddress("bob")
	consumer := chain.PackageAddress(consumerPath)

	testing.SetRealm(testing.NewCodeRealm(tokenPath))
	token, ledger := grc20.NewToken("WrapperAllowance", "WRPL", 4, 0, cur)
	urequire.NoError(t, ledger.Mint(alice, 1_000))
	tokenKey := Register(cross(cur), token, "")

	// No allowance yet: the wrapper cannot touch alice's balance.
	testing.SetRealm(testing.NewCodeRealm(consumerPath))
	uassert.PanicsContains(t, cur, "insufficient allowance", func() {
		TransferFrom(0, cur, tokenKey, alice, bob, 100)
	})
	uassert.Equal(t, int64(1_000), MustGet(tokenKey).BalanceOf(alice))

	// alice approves the consuming realm itself, and only then does it work.
	urequire.NoError(t, ledger.Approve(alice, consumer, 100))

	testing.SetRealm(testing.NewCodeRealm(consumerPath))
	TransferFrom(0, cur, tokenKey, alice, bob, 100)
	uassert.Equal(t, int64(900), MustGet(tokenKey).BalanceOf(alice))
	uassert.Equal(t, int64(100), MustGet(tokenKey).BalanceOf(bob))
}
