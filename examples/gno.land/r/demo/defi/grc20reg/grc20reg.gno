package grc20reg

import (
	"chain"
	"strings"

	"gno.land/p/demo/tokens/grc20"
	"gno.land/p/moul/md"
	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/avl/v0/pager"
	"gno.land/p/nt/avl/v0/rotree"
	"gno.land/p/nt/fqname/v0"
	"gno.land/p/nt/mux/v0"
	"gno.land/p/nt/ufmt/v0"
)

var (
	registry      = avl.NewTree() // rlmPath.symbol -> *Token
	pagerInstance = pager.NewPager(rotree.Wrap(registry, nil), 10, false)
)

// Construction lives in grc20.NewToken — it takes rlm realm last
// and binds origRealm from rlm.PkgPath() under an IsCurrent assertion.
// The registry key is the canonical fqname rlmPath.symbol (one token per
// realm+symbol), independent of Token.ID()'s trailing sequence id, so
// callers can look a token up from the (realm, symbol) pair they already
// know:
//
//	Token, ledger := grc20.NewToken(name, symbol, decimals, id, cur)
//	key := grc20reg.Register(cross(cur), Token, "")

// Register records token under its rlmPath.symbol key and returns that key.
// Token.ID() carries a trailing sequence id (rlmPath.symbol.<id>) that keeps
// token identities/events unique, but the registry deliberately keys by
// rlmPath.symbol so lookups don't need to know the id, and so a realm cannot
// register two tokens under the same symbol (overwrite/alias guard).
func Register(cur realm, token *grc20.Token, slug string) string {
	if token == nil {
		panic("grc20reg: nil token")
	}
	if slug != "" {
		validateSlug(slug)
	}
	rlmPath := cur.Previous().PkgPath()
	key := fqname.Construct(rlmPath, token.GetSymbol())
	// Token.ID() == key + "." + <id>; verify the token originates from the
	// registering realm and symbol.
	if !strings.HasPrefix(token.ID(), key+".") {
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
	router := mux.NewRouter()
	router.HandleFunc("", renderHome)
	router.HandleFunc("*", renderToken)
	return router.Render(path)
}

func renderHome(res *mux.ResponseWriter, req *mux.Request) {
	page := pagerInstance.MustGetPageByPath(req.RawPath)

	if page.TotalItems == 0 {
		res.Write("No registered token.")
		return
	}

	s := ""
	for _, item := range page.Items {
		token := item.Value.(*grc20.Token)
		rlmPath, tokenID := fqname.Parse(item.Key)
		rlmLink := fqname.RenderLink(rlmPath, tokenID)
		infoLink := "/r/demo/defi/grc20reg:" + item.Key
		s += "- " + md.Bold(md.EscapeText(token.GetName())) + " - " + rlmLink + " - " + md.Link("info", infoLink) + "\n"
	}

	s += "\n"
	s += ufmt.Sprintf("**Page %d of %d**\n\n", page.PageNumber, page.TotalPages)
	s += page.Picker(req.Path)
	res.Write(s)
}

func renderToken(res *mux.ResponseWriter, req *mux.Request) {
	key := req.Path
	token := MustGet(key)
	rlmPath, tokenID := fqname.Parse(key)
	rlmLink := fqname.RenderLink(rlmPath, tokenID)
	s := ufmt.Sprintf("# %s\n", md.EscapeText(token.GetName()))
	s += "- symbol: " + md.Bold(md.EscapeText(token.GetSymbol())) + "\n"
	s += ufmt.Sprintf("- realm: %s\n", rlmLink)
	s += ufmt.Sprintf("- decimals: %d\n", token.GetDecimals())
	s += ufmt.Sprintf("- total supply: %d\n", token.TotalSupply())
	res.Write(s)
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
