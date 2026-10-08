package store

import (
	"html/template"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
	"github.com/gnolang/gno/tm2/pkg/bech32"
)

// listing is one app, service or package as the realm reports it. Nothing
// in it is trusted: sanitize validates every field before a template sees it.
type listing struct {
	Slug     string `json:"slug"`
	Kind     string `json:"kind"`
	Path     string `json:"path"`
	Title    string `json:"title"`
	Tagline  string `json:"tagline"`
	Category string `json:"category"`
	Palette  int    `json:"palette"`
	Stars    int    `json:"stars"`  // every star
	Ranked   int    `json:"ranked"` // stars from accounts with a registered name, the ones that rank
	Earned   bool   `json:"earned"`
	Icon     string `json:"icon"`
	Cover    string `json:"cover"`

	// Derived once per cache fill (validCore and sanitize), so a request
	// only reads them: webPath is Path without the domain ("/r/acme/swap").
	webPath   string
	shortPath string
	tier      tier
	art       template.HTML
	glyph     template.HTML
}

// Listing kinds. Only apps are shown on the front page; services and
// packages live on the Build lens.
const (
	kindApp     = "app"
	kindService = "service"
	kindPackage = "package"
)

const (
	maxTitleRunes   = 40
	maxTaglineRunes = 80
	maxEmptyRunes   = 160
	paletteSize     = 12
	ipfsGateway     = "https://ipfs.io/ipfs/" // in gnoweb's CSP img-src
)

var (
	reSlug      = regexp.MustCompile(`^[a-z0-9-]{3,40}$`)
	reNamespace = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,63}$`)
	reKey       = regexp.MustCompile(`^[a-z0-9-]{1,32}$`)
	reIPFS      = regexp.MustCompile(`^ipfs://([A-Za-z0-9]{46,64})$`)
)

// sanitize drops listings whose core fields are invalid and clears rich
// fields that are. Rich fields fail closed: a bad image is not shown, the
// app still is.
func (h *Handler) sanitize(in []listing) []listing {
	out := in[:0]
	for _, l := range in {
		if !h.validCore(&l) {
			h.deps.Logger.Debug("store: dropping invalid listing", "slug", l.Slug)
			continue
		}
		l.Icon, l.Cover = assetURL(l.Icon), assetURL(l.Cover)
		l.shortPath, l.tier = shortPath(l.pkgPath()), h.tierOf(&l)
		l.art, l.glyph = cover(&l), icon(&l)
		out = append(out, l)
	}
	return out
}

// validCore checks identity, path and text.
func (h *Handler) validCore(l *listing) bool {
	web, ok := strings.CutPrefix(l.Path, h.deps.Domain)
	if !ok || !reSlug.MatchString(l.Slug) || !reKey.MatchString(l.Category) {
		return false
	}
	u := weburl.GnoURL{Path: web}
	switch {
	case !u.IsValidPath(),
		l.Kind == kindPackage && !u.IsPure(),
		(l.Kind == kindApp || l.Kind == kindService) && !u.IsRealm(),
		l.Kind != kindApp && l.Kind != kindService && l.Kind != kindPackage:
		return false
	}
	l.webPath = web
	return l.Title != "" && displayable(l.Title, maxTitleRunes) && displayable(l.Tagline, maxTaglineRunes) &&
		l.Palette >= 0 && l.Palette < paletteSize && l.Ranked >= 0 && l.Ranked <= l.Stars
}

// displayable reports whether s is short enough to lay out and free of the
// characters that break or spoof rendering: controls, bidi overrides,
// zero-width marks, blank fillers and doubled spaces. Plain spaces separate
// words and are the only invisible rune kept. Look-alike titles are the
// realm's policy (validText), not a rendering hazard: it alone can keep
// titles unique, and refuses a superset of this.
func displayable(s string, maxRunes int) bool {
	if utf8.RuneCountInString(s) > maxRunes || !utf8.ValidString(s) || strings.Contains(s, "  ") {
		return false
	}
	for _, r := range s {
		if r != ' ' && !components.IsVisibleRune(r) {
			return false
		}
	}
	return true
}

// assetURL maps an owner image to the IPFS gateway, or "" if it is not a
// content id. Only content-addressed images are accepted: a mutable URL
// could change after review, and inline data would bloat every page.
func assetURL(s string) string {
	if m := reIPFS.FindStringSubmatch(s); m != nil {
		return ipfsGateway + m[1]
	}
	return ""
}

// pkgPath is the path without domain and "/r/" or "/p/" ("acme/swap"), the
// form the trust list uses.
func (l *listing) pkgPath() string {
	return l.webPath[len("/r/"):]
}

// namespace returns the first segment of pkgPath.
func (l *listing) namespace() string {
	ns, _, _ := strings.Cut(l.pkgPath(), "/")
	return ns
}

// tier is what a listing may show, decided by gnoweb alone (ADR-004 §2).
type tier int

const (
	tierAnonymous  tier = iota // g1… address namespace
	tierRegistered             // a registered name, not trusted
	tierTrusted                // under the operator's trusted list
)

func (h *Handler) tierOf(l *listing) tier {
	switch {
	case h.deps.Trusted(l.pkgPath()):
		return tierTrusted
	case isAddress(l.namespace()):
		return tierAnonymous
	default:
		return tierRegistered
	}
}

// isAddress reports whether a namespace is a bech32 account address, the
// namespace anyone gets for free.
func isAddress(ns string) bool {
	_, _, err := bech32.Decode(ns)
	return err == nil
}

// rich reports whether owner-supplied assets and the Spotlight are allowed.
func (t tier) rich(earned bool) bool {
	return t == tierTrusted || (t == tierRegistered && earned)
}
