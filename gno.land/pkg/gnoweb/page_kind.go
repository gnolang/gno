package gnoweb

import (
	"strings"

	md "github.com/gnolang/gno/gno.land/pkg/gnoweb/markdown"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
)

// pageKind says who answers for a page, which decides what its <head> may
// repeat and which of its links search engines may follow. gno.land is
// permissionless, so a title, a summary, a share card or a followed link
// lifted from any realm would let anyone speak under gno.land's name (#3910).
// The zero value is the most restrictive, so an unclassified page fails closed.
type pageKind int

const (
	// pageCommunity is a package or user page outside the trusted paths. Its
	// head repeats nothing it renders, and none of its links is followed.
	pageCommunity pageKind = iota
	// pageOfficial is a page of a package under the trusted paths, reached by
	// its own path or through an alias. Its head may repeat its leading
	// heading and paragraph.
	pageOfficial
	// pageOperator is a markdown page an operator passed to -aliases. It is
	// official, and its links are the operator's own.
	pageOperator
	// pageSite is a gnoweb view that belongs to no package, such as the bare
	// "/r/" listing. Its head is generic.
	pageSite
)

const (
	siteDescription             = "Explore realms and packages on gno.land, the network for Gno smart contracts."
	communityRealmDescription   = "A realm deployed on gno.land by its author."
	communityPackageDescription = "A package deployed on gno.land by its author."
	communityUserDescription    = "A gno.land user profile."

	officialImageAsset         = "imgs/og-gnoland.png"
	communityRealmImageAsset   = "imgs/og-community-realm.png"
	communityPackageImageAsset = "imgs/og-community-package.png"
	communityUserImageAsset    = "imgs/og-community-user.png"
)

// pageLead is what a rendered document says about itself: its leading h1
// and its summary. setHeadMetadata decides whether the head repeats it.
type pageLead struct{ title, description string }

// pagePolicy decides, from a page's URL alone, who answers for it and what
// search engines may do with it. It needs no request and no RPC, so anything
// that lists pages, such as a sitemap, classifies them the same way.
type pagePolicy struct {
	trusted trustedPaths
	aliases map[string]AliasTarget
	index   CommunityIndex
}

// kind classifies u, the URL a page is served from once a realm alias is
// resolved. A user page renders that user's home realm. The bare "/r/",
// "/p/" and "/u/" listings belong to no package.
func (p pagePolicy) kind(u *weburl.GnoURL) pageKind {
	if a, ok := p.aliases[u.Path]; ok && a.Kind == StaticMarkdown {
		return pageOperator
	}
	pkg, ok := packagePath(u)
	switch {
	case !ok:
		return pageSite
	case p.trusted.contains(pkg):
		return pageOfficial
	default:
		return pageCommunity
	}
}

// isPackageURL reports whether u is under /r/, /p/ or /u/.
func isPackageURL(u *weburl.GnoURL) bool {
	return u.IsRealm() || u.IsPure() || u.IsUser()
}

// packagePath is the package u renders, without its "/r/" or "/p/" prefix,
// or the user name of a user page. It is false for any other URL, the bare
// listings included.
func packagePath(u *weburl.GnoURL) (string, bool) {
	if !isPackageURL(u) {
		return "", false
	}
	pkg := strings.Trim(u.Path[3:], "/") // skip "/r/", "/p/" or "/u/"
	return pkg, pkg != ""
}

// mayRepeat reports whether the head of u may repeat what the rendered
// document says about itself. Only an official page may, and only when the
// link typed nothing past the path: arguments and query both reach Render,
// and a trusted realm may echo them, as p/gnoland/blog does with a tag in its
// heading. A post's slug and its title share words, so no string match tells
// an echo from a heading.
func (k pageKind) mayRepeat(u *weburl.GnoURL) bool {
	return (k == pageOfficial || k == pageOperator) && u.Args == "" && len(u.Query) == 0
}

// pathTitle names a page by its path alone, for the pages whose document may
// not name them: what the package is and whose namespace it sits in, as in
// "games/chess · realm by nym". An address namespace is shortened the way the
// user page shows it. A path outside /r/, /p/ and /u/ names itself.
func pathTitle(u *weburl.GnoURL) string {
	pkg, ok := packagePath(u)
	if !ok {
		return strings.TrimSuffix(u.Path, "/")
	}
	ns, name, _ := strings.Cut(pkg, "/")
	owner := CreateUsernameFromBech32(ns)
	switch {
	case u.IsUser():
		return owner + " · user profile"
	case u.IsPure() && name == "":
		return "packages by " + owner
	case u.IsPure():
		return name + " · package by " + owner
	case name == "":
		return "realms by " + owner
	default:
		return name + " · realm by " + owner
	}
}

// links says which links of a document of kind k search engines may follow,
// so a page gno.land does not answer for passes on none of its authority.
// External links stay nofollow on every page but an operator's own: a
// trusted realm may still show what its users wrote.
func (k pageKind) links() md.LinkPolicy {
	switch k {
	case pageOperator:
		return md.FollowAllLinks
	case pageCommunity:
		return md.FollowNoLinks
	default:
		return md.FollowInternalLinks
	}
}

// pageRender is what Get hands the views that render a document, and what
// they hand back for the head.
type pageRender struct {
	// links says which links of the document crawlers may follow.
	links md.LinkPolicy
	// markdown is set when the client asked for text/markdown.
	markdown bool
	// lead is set by a view: what the document says about itself.
	lead pageLead
}

// renderContext is the context a document renders under for pr.
func (h *HTTPHandler) renderContext(pr *pageRender) RealmRenderContext {
	return RealmRenderContext{
		ChainId: h.Static.ChainId,
		Remote:  h.Static.RemoteHelp,
		Domain:  h.Static.Domain,
		Links:   pr.links,
	}
}

// card is the summary and share image of a page of kind k at u, for when its
// document may not, or does not, summarise it. gno.land's plain mark goes
// only to the pages it answers for; a community page gets a card that says
// what kind of page it is.
func (k pageKind) card(u *weburl.GnoURL) (description, image string) {
	switch {
	case k != pageCommunity:
		return siteDescription, officialImageAsset
	case u.IsPure():
		return communityPackageDescription, communityPackageImageAsset
	case u.IsUser():
		return communityUserDescription, communityUserImageAsset
	default:
		return communityRealmDescription, communityRealmImageAsset
	}
}
