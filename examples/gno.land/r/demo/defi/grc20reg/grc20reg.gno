package grc20reg

import (
	"chain"
	"errors"
	"strings"

	"gno.land/p/demo/tokens/grc20"
	"gno.land/p/moul/md"
	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/avl/v0/rotree"
	"gno.land/p/nt/fqname/v0"
	"gno.land/p/nt/ufmt/v0"
)

var registry = avl.NewTree() // rlmPath.slug -> *Token

// Construction lives in grc20.NewToken — it takes rlm realm last
// and binds origRealm from rlm.PkgPath() under an IsCurrent assertion.
// Token.ID() is the token identity (rlmPath.symbol.id). The registry slug is
// a caller-chosen lookup alias, independent of the token's identity:
//
//	Token, ledger := grc20.NewToken(name, symbol, decimals, id, cur)
//	key := grc20reg.Register(cross(cur), Token, "my-token")

// Register records token under rlmPath.slug and returns that registry key.
// The registering realm is responsible for allocating unique Token IDs and
// for deciding whether one token should be registered under multiple aliases.
func Register(cur realm, token *grc20.Token, slug string) string {
	if token == nil {
		panic("grc20reg: nil token")
	}
	// Reject stale or forged realm frames before trusting cur.Previous() as
	// the registering realm.
	if !cur.IsCurrent() {
		panic("grc20reg: rlm is not the caller's live cur")
	}
	if slug == "" {
		panic("grc20reg: empty slug")
	}
	if err := validateSlug(slug); err != nil {
		panic(err.Error())
	}
	rlmPath := cur.Previous().PkgPath()
	key := fqname.Construct(rlmPath, slug)
	tokenIDPrefix := fqname.Construct(rlmPath, token.GetSymbol()) + "."
	if !strings.HasPrefix(token.ID(), tokenIDPrefix) {
		panic("grc20reg: token must be registered from its own realm")
	}
	if registry.Has(key) {
		panic("grc20reg: token already registered")
	}
	registry.Set(key, token)
	chain.Emit(
		registerEvent,
		"token_path", key,
		"pkgpath", rlmPath,
		"slug", slug,
		"symbol", token.GetSymbol(),
		"token_id", token.ID(),
	)
	return key
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

// validateSlug rejects slugs that are too long or contain non-alphanumeric characters.
// Only letters, digits, dashes, and underscores are allowed.
func validateSlug(slug string) error {
	if len(slug) > maxSlugLen {
		return errors.New("grc20reg: slug too long")
	}
	for _, c := range slug {
		if !isAlphanumeric(c) && c != '_' && c != '-' {
			return errors.New("grc20reg: invalid slug character: " + string(c))
		}
	}
	return nil
}

func isAlphanumeric(c rune) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}
