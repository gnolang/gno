package grc20reg

import (
	"chain"
	"strings"

	"gno.land/p/demo/tokens/grc20"
	"gno.land/p/moul/md"
	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/avl/v0/rotree"
	"gno.land/p/nt/fqname/v0"
	"gno.land/p/nt/ufmt/v0"
	"gno.land/p/onbloc/identifier"
)

var (
	registry    = avl.NewTree() // tokenKey (fqname rlmPath.slug) -> *Token
	idGenerator *identifier.Generator
)

func init(cur realm) {
	idGenerator = identifier.NewGenerator(0, cur)
}

// IdentifierGenerator returns the registry's canonical id generator.
// Pass it to grc20.NewToken so the resulting id is issued by grc20reg and accepted by Register.
func IdentifierGenerator() *identifier.Generator {
	return idGenerator
}

// Register records token under a tokenKey (fqname rlmPath.slug) and returns
// that key. The tokenKey is a human-chosen alias for lookup and is a distinct
// value from Token.ID() — the unforgeable, generator-issued identifier emitted
// in the token's events. A realm may register one token under several slugs
// (aliases); re-registering the same slug is rejected. The register event
// carries both so consumers can map tokenKey <-> tokenID:
//
//	Token, ledger := grc20.NewToken(name, symbol, decimals, grc20reg.IdentifierGenerator(), cur)
//	key := grc20reg.Register(cross(cur), Token, "myslug") // key == rlmPath.myslug
func Register(cur realm, token *grc20.Token, slug string) string {
	if token == nil {
		panic("grc20reg: nil token")
	}

	if slug != "" {
		validateSlug(slug)
	}

	rlmPath := cur.Previous().PkgPath()

	// Token.ID() == rlmPath + "." + symbol + "." + <GeneratorID>;
	// verify the token originates from the registering realm.
	if !strings.HasPrefix(token.ID(), rlmPath+".") {
		panic("grc20reg: token must be registered from its own realm")
	}

	// verify the token id must have been issued by the registry's own generator.
	if token.IdentifierPath() != idGenerator.PackagePath() {
		panic("grc20reg: token id not issued by the registry generator")
	}

	// construct the token key (fqname rlmPath.slug)
	tokenKey := fqname.Construct(rlmPath, slug)
	if registry.Has(tokenKey) {
		panic("grc20reg: token already registered")
	}

	registry.Set(tokenKey, token)

	chain.Emit(
		registerEvent,
		"token_key", tokenKey,
		"token_id", token.ID(),
		"pkgpath", rlmPath,
		"slug", slug,
		"symbol", token.GetSymbol(),
	)

	return tokenKey
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

func Render(path string) string {
	switch {
	case path == "": // home
		// TODO: add pagination
		s := ""
		count := 0
		registry.Iterate("", "", func(key string, tokenI any) bool {
			count++
			token := tokenI.(*grc20.Token)
			rlmPath, tokenID := fqname.Parse(key)
			rlmLink := fqname.RenderLink(rlmPath, tokenID)
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
		rlmPath, tokenID := fqname.Parse(key)
		rlmLink := fqname.RenderLink(rlmPath, tokenID)
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
