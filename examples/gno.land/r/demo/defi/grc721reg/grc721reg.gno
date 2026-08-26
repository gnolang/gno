// Package grc721reg discovers GRC721 collections by fqname (rlmPath.slug).
// Open to any extension via read-only ExtensionView; no concrete extension imports.
//
// Unlike grc20reg, which stores a concrete *grc20.Token, this registry stores the
// caller-supplied ExtensionView interface values and hands them back out, so a
// consumer must type-assert an ExtensionView to a concrete view rather than trust the
// interface. That is a deliberate trade: it keeps the registry open to extension kinds
// it does not import, at the cost of moving the type check to the consumer.
package grc721reg

import (
	"chain"
	"strings"

	"gno.land/p/demo/tokens/grc721"
	"gno.land/p/demo/tokens/grc721/collection"
	"gno.land/p/moul/md"
	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/avl/v0/rotree"
	"gno.land/p/nt/fqname/v0"
	"gno.land/p/nt/ufmt/v0"
)

var registry = avl.NewTree() // fqname(rlmPath, slug); empty slug allowed

// Register keys the collection under rlmPath.slug. Token must originate from the
// registering realm (Token.ID prefix check).
func Register(cur realm, token *grc721.Token, slug string, extensions ...grc721.ExtensionView) string {
	if token == nil {
		panic("grc721reg: nil token")
	}

	if slug != "" {
		validateSlug(slug)
	}

	// The registering realm is derived from the live crossing frame, so verify
	// the frame before reading it: .Previous() on a stale token resolves toward
	// the origin instead of failing, which would key the entry off the wrong
	// identity.
	if !cur.IsCurrent() {
		panic(grc721.ErrSpoofedRealm)
	}

	rlmPath := cur.Previous().PkgPath()
	if !strings.HasPrefix(token.ID(), rlmPath+".") {
		panic("grc721reg: collection must be registered from its own realm")
	}

	key := fqname.Construct(rlmPath, slug)
	if registry.Has(key) {
		panic("grc721reg: collection already registered")
	}

	col := collection.NewCollection(token, extensions...)

	registry.Set(key, col)
	chain.Emit(
		RegisterEvent,
		"token_key", key,
		"token_id", token.ID(),
		"pkgpath", rlmPath,
		"slug", slug,
		"symbol", token.GetSymbol(),
		"extensions", strings.Join(col.Kinds(), ","),
	)

	return key
}

func Get(key string) *collection.Collection {
	v := registry.Get(key)
	if v == nil {
		return nil
	}

	return v.(*collection.Collection)
}

func MustGet(key string) *collection.Collection {
	v := Get(key)
	if v == nil {
		panic("grc721reg: unknown collection: " + key)
	}

	return v
}

func GetToken(key string) *grc721.Token {
	v := Get(key)
	if v == nil {
		return nil
	}

	return v.Token()
}

func GetExtension(key, kind string) (grc721.ExtensionView, bool) {
	v := Get(key)
	if v == nil {
		return nil, false
	}

	return v.Extension(kind)
}

func GetRegistry() *rotree.ReadOnlyTree {
	return rotree.Wrap(registry, nil)
}

func Render(path string) string {
	switch {
	case path == "":
		s := ""
		count := 0

		registry.Iterate("", "", func(key string, vI any) bool {
			count++

			v := vI.(*collection.Collection)
			rlmPath, sym := fqname.Parse(key)
			rlmLink := fqname.RenderLink(rlmPath, sym)
			infoLink := "/r/demo/defi/grc721reg:" + key
			s += "- " + md.Bold(md.EscapeText(v.Token().GetName())) + " - " + rlmLink +
				" - " + extensionBadges(v) + " - " + md.Link("info", infoLink) + "\n"

			return false
		})

		if count == 0 {
			return "No registered collection."
		}

		return s
	default:
		key := path
		// Render takes attacker-controlled input, so an unknown key renders a
		// message rather than aborting the page: MustGet would let any visitor
		// break this view with a bad URL.
		v := Get(key)
		if v == nil {
			return "Unknown collection: " + md.EscapeText(key)
		}
		rlmPath, sym := fqname.Parse(key)
		rlmLink := fqname.RenderLink(rlmPath, sym)

		s := ufmt.Sprintf("# %s\n", md.EscapeText(v.Token().GetName()))
		s += "- symbol: " + md.Bold(md.EscapeText(v.Token().GetSymbol())) + "\n"
		s += ufmt.Sprintf("- realm: %s\n", rlmLink)
		s += ufmt.Sprintf("- total supply: %d\n", v.Token().TotalSupply())
		s += "- extensions: " + extensionBadges(v) + "\n"

		return s
	}
}

// extensionBadges renders the kinds a collection advertises. ExtensionKind() is
// issuer-supplied, so it goes through md.InlineCode rather than being pasted
// between literal backticks: a backtick in a kind would otherwise close the
// code span and inject markdown into the public index. InlineCode picks a
// delimiter wide enough to outscan the content, which hand-escaping cannot do
// inside a code span.
func extensionBadges(v *collection.Collection) string {
	kinds := v.Kinds()
	if len(kinds) == 0 {
		return "_core only_"
	}

	out := ""

	for _, k := range kinds {
		out += md.InlineCode(k) + " "
	}

	return strings.TrimSpace(out)
}

func validateSlug(slug string) {
	if len(slug) > maxSlugLen {
		panic("grc721reg: slug too long")
	}

	for _, c := range slug {
		if !isAlphanumeric(c) && c != '_' && c != '-' {
			panic("grc721reg: invalid slug character: " + string(c))
		}
	}
}

func isAlphanumeric(c rune) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}
