package grc20reg

import (
	"chain"
	"chain/runtime"
	"strings"

	"gno.land/p/moul/md/v0"
	"gno.land/p/nt/avl/pager/v0"
	"gno.land/p/nt/avl/rotree/v0"
	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/fqname/v0"
	"gno.land/p/nt/grc20/v0"
	"gno.land/p/nt/mux/v0"
	"gno.land/p/nt/ufmt/v0"
)

var (
	registry      = avl.NewTree() // rlmPath.symbol -> *handle
	pagerInstance = pager.NewPager(rotree.Wrap(registry, nil), 10, false)
)

// handle is the registry's own record: the token (read-only for consumers)
// plus the token realm's crossing functions, which are the only way to move it.
// A realm cannot construct another realm's struct type, so the registry builds
// this value itself from the functions a registering realm hands in.
type handle struct {
	token        *grc20.Token
	transfer     func(cur realm, to address, amount int64)
	approve      func(cur realm, spender address, amount int64)
	transferFrom func(cur realm, from, to address, amount int64)
}

// Construction lives in grc20.NewToken — it takes rlm realm last
// and binds origRealm from rlm.PkgPath() under an IsCurrent assertion.
// The registry key is the canonical fqname rlmPath.symbol (one token per
// realm+symbol), independent of Token.ID()'s trailing sequence id, so
// callers can look a token up from the (realm, symbol) pair they already
// know:
//
//	Token, ledger := grc20.NewToken(name, symbol, decimals, id, cur)
//	key := grc20reg.Register(cross(cur), Token, "", Transfer, Approve, TransferFrom)

// Register records token under its rlmPath.symbol key and returns that key.
// Token.ID() carries a trailing sequence id (rlmPath.symbol.<id>) that keeps
// token identities/events unique, but the registry deliberately keys by
// rlmPath.symbol so lookups don't need to know the id, and so a realm cannot
// register two tokens under the same symbol (overwrite/alias guard).
func Register(cur realm, token *grc20.Token, slug string,
	transfer func(cur realm, to address, amount int64),
	approve func(cur realm, spender address, amount int64),
	transferFrom func(cur realm, from, to address, amount int64),
) string {
	if token == nil {
		panic("grc20reg: nil token")
	}
	if transfer == nil || approve == nil || transferFrom == nil {
		panic("grc20reg: nil crossing function")
	}
	if slug != "" {
		validateSlug(slug)
	}
	rlmPath := cur.Previous().PkgPath()
	// A run script's functions do not outlive the transaction that defines them.
	if cur.Previous().IsEphemeral() {
		panic("grc20reg: a run script cannot register a token")
	}
	key := fqname.Construct(rlmPath, token.GetSymbol())
	// Token.ID() == key + "." + <id>; verify the token originates from the
	// registering realm and symbol.
	if !strings.HasPrefix(token.ID(), key+".") {
		panic("grc20reg: token must be registered from its own realm")
	}
	// The registry vouches for whose code runs, not for what it does: every
	// function value carries the package that declares it, so a realm cannot
	// register another realm's move beside its own token.
	if runtime.FuncPkgPath(transfer) != rlmPath ||
		runtime.FuncPkgPath(approve) != rlmPath ||
		runtime.FuncPkgPath(transferFrom) != rlmPath {
		panic("grc20reg: functions must be declared in the registering realm")
	}
	if registry.Has(key) {
		panic("grc20reg: token already registered")
	}
	// Keep a copy: the registrant can overwrite its own token struct later
	// (`*Token = *other.Token`) and would make this key name another ledger.
	tok := *token
	registry.Set(key, &handle{&tok, transfer, approve, transferFrom})
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
	h := getHandle(key)
	if h == nil {
		return nil
	}
	return h.token
}

func getHandle(key string) *handle {
	h := registry.Get(key)
	if h == nil {
		return nil
	}
	return h.(*handle)
}

func MustGet(key string) *grc20.Token {
	token := Get(key)
	if token == nil {
		panic("unknown token: " + key)
	}
	return token
}

func mustGetHandle(key string) *handle {
	h := getHandle(key)
	if h == nil {
		panic("grc20reg: unknown token: " + key)
	}
	return h
}

// Write wrappers: non-crossing, so the caller forwards its own live cur and the
// stored token-realm crossing function sees the caller as cur.Previous(); the
// actor is the calling realm and the registry never spends its own balance.

// Transfer moves `amount` out of the CALLING REALM's own balance.
func Transfer(_ int, rlm realm, tokenKey string, to address, amount int64) {
	mustGetHandle(tokenKey).transfer(cross(rlm), to, amount)
}

// Approve sets an allowance owned by the CALLING REALM.
func Approve(_ int, rlm realm, tokenKey string, spender address, amount int64) {
	mustGetHandle(tokenKey).approve(cross(rlm), spender, amount)
}

// TransferFrom spends an allowance with the CALLING REALM as the spender.
func TransferFrom(_ int, rlm realm, tokenKey string, from, to address, amount int64) {
	mustGetHandle(tokenKey).transferFrom(cross(rlm), from, to, amount)
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
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
		token := item.Value.(*handle).token
		rlmPath, tokenID := fqname.Parse(item.Key)
		rlmLink := fqname.RenderLink(rlmPath, tokenID)
		infoLink := "/r/nt/grc20reg/v0:" + item.Key
		s += "- " + md.Bold(md.EscapeText(token.GetName())) + " - " + rlmLink + " - " + md.Link("info", infoLink) + "\n"
	}

	s += "\n"
	s += ufmt.Sprintf("**Page %d of %d**\n\n", page.PageNumber, page.TotalPages)
	s += page.Picker(req.RawPath)
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

// GetRegistry exposes a read-only view whose values are *grc20.Token; the
// registry's handle stays private.
func GetRegistry() *rotree.ReadOnlyTree {
	return rotree.Wrap(registry, func(v any) any {
		return v.(*handle).token
	})
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
