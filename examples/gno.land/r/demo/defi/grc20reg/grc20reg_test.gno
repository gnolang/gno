package grc20reg

import (
	"strings"
	"testing"

	"gno.land/p/demo/tokens/grc20"
	"gno.land/p/nt/urequire/v0"
	"gno.land/p/onbloc/identifier"
)

func TestRegistry(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/foo"))
	token, ledger := grc20.NewToken("TestToken", "TST", 4, IdentifierGenerator(), cur)
	ledger.Mint(cur.Address(), 1234567)
	// The returned tokenKey is the slug alias, distinct from Token.ID().
	key := Register(cross(cur), token, "mySlug")
	urequire.Equal(t, "gno.land/r/demo/foo.mySlug", key)
	urequire.NotEqual(t, token.ID(), key)
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
}

func TestRegisterAliases(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/grc20reg_alias"))
	token, _ := grc20.NewToken("Aliased", "ALIAS", 4, IdentifierGenerator(), cur)

	// The same token may be registered under several slugs (aliases).
	keyA := Register(cross(cur), token, "a")
	keyB := Register(cross(cur), token, "b")
	urequire.NotEqual(t, keyA, keyB)
	urequire.Equal(t, token.ID(), Get(keyA).ID())
	urequire.Equal(t, token.ID(), Get(keyB).ID())
}

func TestRegisterRejectsDuplicateSlug(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/grc20reg_dup"))
	first, _ := grc20.NewToken("First", "FIR", 4, IdentifierGenerator(), cur)
	second, _ := grc20.NewToken("Second", "SEC", 4, IdentifierGenerator(), cur)
	Register(cross(cur), first, "dup")

	// The same slug (tokenKey) is rejected, even for a different token.
	urequire.AbortsContains(t, cur, "token already registered", func() {
		Register(cross(cur), second, "dup")
	})
}

func TestRegisterRejectsForeignGenerator(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/grc20reg_foreign_gen"))
	// A token whose id came from a generator other than the registry's own is
	// rejected — this is what ties a registered token to the trusted generator.
	ownGen := identifier.NewGenerator(0, cur)
	token, _ := grc20.NewToken("Foreign", "FGN", 4, ownGen, cur)

	urequire.AbortsContains(t, cur, "not issued by the registry generator", func() {
		Register(cross(cur), token, "")
	})
}

func TestRegisterRejectsTokenFromDifferentRealm(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/grc20reg_id_source"))
	gen := IdentifierGenerator()
	token, _ := grc20.NewToken("Mismatch Token", "MISMATCH", 4, gen, cur)

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
