package grc20reg

import (
	"chain"

	"gno.land/p/demo/tokens/grc20"
	"gno.land/p/moul/md"
	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/avl/v0/rotree"
	"gno.land/p/nt/fqname/v0"
	"gno.land/p/nt/ufmt/v0"
)

var registry = avl.NewTree() // rlmPath[.slug] -> *Token (slug is optional)

// Construction lives in grc20.NewToken — it now takes (_ int, rlm realm)
// and binds origRealm from rlm.PkgPath() under an IsCurrent assertion.
// Use this package's Register to make a token discoverable:
//
//	Token, ledger := grc20.NewToken(0, cur, name, symbol, decimals)
//	grc20reg.Register(cross, Token, "")

func Register(cur realm, token *grc20.Token, slug string) {
	assertCurrent(cur)
	if slug != "" {
		validateSlug(slug)
	}
	rlmPath := cur.Previous().PkgPath()
	key := fqname.Construct(rlmPath, slug)
	registry.Set(key, token)
	chain.Emit(
		registerEvent,
		"pkgpath", rlmPath,
		"slug", slug,
	)
}

func Get(key string) *grc20.Token {
	token := registry.Get(key)
	if token == nil {
		return nil
	}
	return token.(*grc20.Token)
}

func MustGet(key string) *grc20.Token {
	token := Get(key)
	if token == nil {
		panic("unknown token: " + key)
	}
	return token
}

// Transfer moves tokens owned by the immediate caller. A direct user call
// spends the user's balance; a realm cross-call spends the calling realm's.
func Transfer(cur realm, tokenKey string, to address, amount int64) {
	assertCurrent(cur)
	checkErr(MustGet(tokenKey).CallerTeller().Transfer(0, cur, to, amount))
}

// Approve sets an allowance owned by the immediate caller. A direct user call
// uses the user's allowance; a realm cross-call uses the calling realm's.
func Approve(cur realm, tokenKey string, spender address, amount int64) {
	assertCurrent(cur)
	checkErr(MustGet(tokenKey).CallerTeller().Approve(0, cur, spender, amount))
}

// TransferFrom spends an allowance as the immediate caller. A direct user is
// the spender; for a realm cross-call, the calling realm is the spender.
func TransferFrom(cur realm, tokenKey string, from, to address, amount int64) {
	assertCurrent(cur)
	checkErr(MustGet(tokenKey).CallerTeller().TransferFrom(0, cur, from, to, amount))
}

func Render(path string) string {
	switch {
	case path == "": // home
		// TODO: add pagination
		s := ""
		count := 0
		registry.Iterate("", "", func(key string, tokenI any) bool {
			count++
			token := tokenI.(*grc20.Token)
			rlmPath, slug := fqname.Parse(key)
			rlmLink := fqname.RenderLink(rlmPath, slug)
			infoLink := "/r/demo/grc20reg:" + key
			s += "- " + md.Bold(md.EscapeText(token.GetName())) + " - " + rlmLink + " - " + md.Link("info", infoLink) + "\n"
			return false
		})
		if count == 0 {
			return "No registered token."
		}
		return s
	default: // specific token
		key := path
		token := MustGet(key)
		rlmPath, slug := fqname.Parse(key)
		rlmLink := fqname.RenderLink(rlmPath, slug)
		s := ufmt.Sprintf("# %s\n", md.EscapeText(token.GetName()))
		s += "- symbol: " + md.Bold(md.EscapeText(token.GetSymbol())) + "\n"
		s += ufmt.Sprintf("- realm: %s\n", rlmLink)
		s += ufmt.Sprintf("- decimals: %d\n", token.GetDecimals())
		s += ufmt.Sprintf("- total supply: %d\n", token.TotalSupply())
		return s
	}
}

const (
	registerEvent = "register"
	maxSlugLen    = 128
)

func GetRegistry() *rotree.ReadOnlyTree {
	return rotree.Wrap(registry, nil)
}

func assertCurrent(cur realm) {
	if !cur.IsCurrent() {
		panic(grc20.ErrSpoofedRealm)
	}
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

// validateSlug panics if the slug is too long or contains non-alphanumeric characters.
// Only letters, digits, dashes, and underscores are allowed.
func validateSlug(slug string) {
	if len(slug) > maxSlugLen {
		panic("grc20reg: slug too long")
	}
	for _, c := range slug {
		if !isAlphanumeric(c) && c != '_' && c != '-' {
			panic("grc20reg: invalid slug character: " + string(c))
		}
	}
}

func isAlphanumeric(c rune) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}
