package grc20reg

import (
	"chain"
	"strings"

	"gno.land/p/demo/tokens/grc20"
	"gno.land/p/moul/md"
	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/avl/v0/rotree"
	"gno.land/p/nt/fqname/v0"
	"gno.land/p/nt/seqid/v0"
	"gno.land/p/nt/ufmt/v0"
)

var (
	registry = avl.NewTree() // fqname rlmPath.symbol -> *Token

	idSeq seqid.ID // id counter; only nextIDSeq, a closure over it, leaves this realm
)

// IdentifierGenerator returns a generator backed by the registry's counter.
// Tokens built with it carry grc20reg as their IdentifierPath(), which is what
// Register requires:
//
//	gen := grc20reg.IdentifierGenerator(cross(cur))
//	Token, ledger := grc20.NewToken(name, symbol, decimals, gen, cur)
//	key := grc20reg.Register(cross(cur), Token, "")
//
// Holding a generator grants no more than calling this again would — it draws
// from the counter without revealing or forking it.
func IdentifierGenerator(cur realm) *grc20.IDGenerator {
	return grc20.NewIDGenerator(0, cur, nextIDSeq)
}

// nextIDSeq advances the registry's counter. Declared in this realm, so the
// declaring-realm borrow gives it authority over idSeq wherever it is called.
func nextIDSeq() string {
	return idSeq.Next().String()
}

// Register records token under its rlmPath.symbol key and returns that key.
// Token.ID() carries a trailing generated id (rlmPath.symbol.<id>) that keeps
// token identities/events unique, but the registry deliberately keys by
// rlmPath.symbol so lookups don't need to know the id, and so a realm cannot
// register two tokens under the same symbol (overwrite/alias guard). slug is
// recorded in the register event only.
//
// The id must have come from IdentifierGenerator; one produced anywhere else is
// not from this counter and could collide with one that is.
func Register(cur realm, token *grc20.Token, slug string) string {
	if token == nil {
		panic("grc20reg: nil token")
	}
	if slug != "" {
		validateSlug(slug)
	}
	rlmPath := cur.Previous().PkgPath()
	key := fqname.Construct(rlmPath, token.GetSymbol())
	// Token.ID() == key + "." + <id>; verify the token originates from
	// the registering realm and symbol.
	if !strings.HasPrefix(token.ID(), key+".") {
		panic("grc20reg: token must be registered from its own realm")
	}
	// PackagePath() is bound under IsCurrent, so no other realm can create a
	// generator claiming this one.
	if token.IdentifierPath() != cur.PkgPath() {
		panic("grc20reg: token id not issued by the registry generator")
	}
	if registry.Has(key) {
		panic("grc20reg: token already registered")
	}
	registry.Set(key, token)
	chain.Emit(
		registerEvent,
		"token_path", key,
		"token_id", token.ID(),
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

// Transfer moves tokens owned by the immediate caller. A direct user call
// spends the user's balance. A realm cross-call spends the calling realm's
// balance.
//
// This differs from calling `Get(tokenKey).CallerTeller().Transfer` within that
// realm, which spends the balance of the realm's previous caller. To act as
// your own realm explicitly, use `Get(tokenKey).RealmTeller(0, cur)`.
func Transfer(cur realm, tokenKey string, to address, amount int64) {
	checkErr(MustGet(tokenKey).CallerTeller().Transfer(0, cur, to, amount))
}

// Approve sets an allowance owned by the immediate caller. A direct user call
// updates the user's allowance. A realm cross-call updates the calling realm's
// allowance.
//
// This differs from calling `Get(tokenKey).CallerTeller().Approve` within that
// realm, which updates the allowance of the realm's previous caller. To act as
// your own realm explicitly, use `Get(tokenKey).RealmTeller(0, cur)`.
func Approve(cur realm, tokenKey string, spender address, amount int64) {
	checkErr(MustGet(tokenKey).CallerTeller().Approve(0, cur, spender, amount))
}

// TransferFrom spends an allowance as the immediate caller. A direct user call
// uses the user as the spender. A realm cross-call uses the calling realm as
// the spender.
//
// This differs from calling `Get(tokenKey).CallerTeller().TransferFrom` within
// that realm, which uses the realm's previous caller as the spender. To act as
// your own realm explicitly, use `Get(tokenKey).RealmTeller(0, cur)`.
func TransferFrom(cur realm, tokenKey string, from, to address, amount int64) {
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
