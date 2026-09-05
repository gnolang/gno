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

// Write wrappers: a registered token can be moved through the registry without
// importing the token's realm, which is the point of a registry. What makes
// that safe is the calling convention, so it is worth stating once here rather
// than three times below.
//
// These are NOT crossing functions. `_ int, rlm realm` is the only shape that
// gives a non-crossing realm parameter — a realm parameter in first position
// must be named `cur`, which makes the function crossing — and the distinction
// is load-bearing, not stylistic:
//
//   - Crossing (`func Transfer(cur realm, …)`) mints a fresh `cur` for THIS
//     realm. RealmTeller would then bind the actor to the registry's own
//     address and the registry would spend its own balance. Useless at best.
//   - Non-crossing (`func Transfer(_ int, rlm realm, …)`) declaring-borrows to
//     the registry without a realm-context change, so `rlm` is still the
//     caller's own live token and the actor is the caller.
//
// The safety comes from RealmTeller's IsCurrent() assertion. The actor is
// rlm.Address() on a token that must be the live crossing frame, so it is
// provably the immediate caller: a stale or foreign token is refused with
// ErrSpoofedRealm. Debiting anyone else would mean holding their live `cur`,
// which means executing inside their frame — authority they handed over
// deliberately, and the same trust model RealmTeller already carries.
//
// This is deliberately not grc20.CallerTeller. "Act as whoever called me" is
// the confused deputy: the debited account ends up chosen by whoever the hub
// can be induced to serve, and an intermediate realm frame silently changes who
// pays. CallerTeller is confined to the token's own realm for that reason and
// is not reachable from a *Token. "Act as the realm that called me, verified
// current" has nothing to induce — the caller cannot name a victim, only
// itself.
//
// Realm-only by construction: MsgCall cannot build a realm argument
// (convertArgToGno rejects non-primitive parameter types), so a signing user
// cannot reach these at all and there is no in-band case to guard against.
// Users move their own tokens through the token realm's own entry points
// (wugnot.Transfer, foo20.Transfer, …).

// Transfer moves `amount` out of the CALLING REALM's own balance.
//
// Call it non-crossing, forwarding your own `cur`:
//
//	grc20reg.Transfer(0, cur, "gno.land/r/demo/defi/foo20.FOO", to, 100)
func Transfer(_ int, rlm realm, tokenKey string, to address, amount int64) {
	checkErr(MustGet(tokenKey).RealmTeller(0, rlm).Transfer(0, rlm, to, amount))
}

// Approve sets an allowance owned by the CALLING REALM, letting `spender` draw
// on the calling realm's balance. It does not touch the signing user's
// allowances.
func Approve(_ int, rlm realm, tokenKey string, spender address, amount int64) {
	checkErr(MustGet(tokenKey).RealmTeller(0, rlm).Approve(0, rlm, spender, amount))
}

// TransferFrom spends an allowance with the CALLING REALM as the spender.
//
// Note the allowance direction this implies: `from` must have approved the
// calling realm, not the signing user. That is the supported way for a realm to
// move a user's funds — the user grants the realm an allowance, and the realm
// draws on it as itself, so every debit is one the owner authorized against
// that specific realm.
func TransferFrom(_ int, rlm realm, tokenKey string, from, to address, amount int64) {
	checkErr(MustGet(tokenKey).RealmTeller(0, rlm).TransferFrom(0, rlm, from, to, amount))
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
