// Package grc721reg discovers GRC721 collections by fqname (rlmPath.symbol).
// Open to any extension via read-only ExtensionView; no concrete extension imports.
//
// Unlike grc20reg, which stores a concrete *grc20.Token, this registry stores the
// caller's *collection.Collection and hands its ExtensionView values back out, so a
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
	"gno.land/p/moul/realmpath"
	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/avl/v0/pager"
	"gno.land/p/nt/avl/v0/rotree"
	"gno.land/p/nt/fqname/v0"
	"gno.land/p/nt/ufmt/v0"
)

var registry = avl.NewTree() // fqname(rlmPath, symbol); one collection per realm+symbol

// Register keys the collection under rlmPath.symbol. Token must originate from
// the registering realm (Token.ID prefix check).
//
// The caller's live *collection.Collection is stored rather than a copy, so an
// extension attached after registration still reaches consumers.
func Register(cur realm, col *collection.Collection, slug string) string {
	if col == nil {
		panic("grc721reg: nil collection")
	}

	token := col.Token()

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

	prev := cur.Previous()
	if prev.IsUser() {
		panic("grc721reg: registration requires a persistent realm")
	}

	rlmPath := prev.PkgPath()
	// Token.ID() == key + "." + <id>. Keying on the token's own symbol instead
	// of the caller-chosen slug verifies realm and symbol in one check, and
	// holds a collection to a single entry whose suffix is the real symbol.
	key := fqname.Construct(rlmPath, token.GetSymbol())
	if !strings.HasPrefix(token.ID(), key+".") {
		panic("grc721reg: collection must be registered from its own realm")
	}

	if registry.Has(key) {
		panic("grc721reg: collection already registered")
	}

	attrs := []string{
		"token_path", key,
		"token_id", token.ID(),
		"pkgpath", rlmPath,
		"slug", slug,
		"symbol", token.GetSymbol(),
	}

	// One attribute per kind: a joined list splits wrongly on a kind that
	// carries the separator. Both bounds are checked here so an oversized set
	// fails with the registry's name on it rather than the event's.
	kinds := col.KindsUpTo(maxEventKinds + 1)
	if len(kinds) > maxEventKinds {
		panic("grc721reg: too many extensions")
	}

	for _, k := range kinds {
		if len(k) > maxKindLen {
			panic("grc721reg: extension kind too long")
		}

		attrs = append(attrs, "extension", k)
	}

	registry.Set(key, col)
	chain.Emit(RegisterEvent, attrs...)

	return key
}

// Unregister drops the calling realm's own entry, so an index that only ever
// grows can also shrink.
func Unregister(cur realm, symbol string) string {
	if !cur.IsCurrent() {
		panic(grc721.ErrSpoofedRealm)
	}

	rlmPath := cur.Previous().PkgPath()
	key := fqname.Construct(rlmPath, symbol)

	if _, removed := registry.Remove(key); !removed {
		panic("grc721reg: collection not registered")
	}

	chain.Emit(
		UnregisterEvent,
		"token_path", key,
		"pkgpath", rlmPath,
		"symbol", symbol,
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
	req := realmpath.Parse(path)
	if req.Path == "" {
		return renderIndex(path)
	}

	return renderCollection(req.Path)
}

// renderIndex pages the walk. Registration is open to any realm and an entry
// costs the page every visitor reads, so the number of entries one render
// touches must not be the registrants' to choose. SizeQueryParam is left unset
// for the same reason: a visitor cannot widen the page either.
func renderIndex(path string) string {
	p := &pager.Pager{
		Tree:            GetRegistry(),
		PageQueryParam:  "page",
		DefaultPageSize: pageSize,
	}

	page, err := p.GetPageByPath(path)
	if err != nil {
		page = p.GetPage(1)
	}

	if len(page.Items) == 0 && page.TotalItems > 0 {
		page = p.GetPage(1)
	}

	if len(page.Items) == 0 {
		return "No registered collection."
	}

	s := ""

	for _, item := range page.Items {
		v := item.Value.(*collection.Collection)
		rlmPath, sym := fqname.Parse(item.Key)
		rlmLink := fqname.RenderLink(rlmPath, sym)
		infoLink := "/r/demo/defi/grc721reg:" + item.Key
		s += "- " + md.Bold(md.EscapeText(v.Token().GetName())) + " - " + rlmLink +
			" - " + extensionBadges(v) + " - " + md.Link("info", infoLink) + "\n"
	}

	// Built from the page number alone, so no other query the visitor supplied
	// is carried into the links.
	if picker := page.Picker(""); picker != "" {
		s += "\n" + picker + "\n"
	}

	return s
}

func renderCollection(key string) string {
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

// extensionBadges renders the kinds a collection advertises. ExtensionKind() is
// issuer-supplied, so it goes through md.InlineCode rather than being pasted
// between literal backticks: a backtick in a kind would otherwise close the
// code span and inject markdown into the public index. InlineCode picks a
// delimiter wide enough to outscan the content, which hand-escaping cannot do
// inside a code span.
func extensionBadges(v *collection.Collection) string {
	kinds := v.KindsUpTo(maxBadges + 1)
	if len(kinds) == 0 {
		return "_core only_"
	}

	// The set stays live and a kind is issuer-supplied, so neither its count
	// nor its length is this page's to trust.
	out := ""

	for i, k := range kinds {
		if i == maxBadges {
			out += "…"
			break
		}

		if len(k) > maxKindLen {
			k = strings.ToValidUTF8(k[:maxKindLen], "") + "…"
		}

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
