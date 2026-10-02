package gnoweb

import (
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
	// pageOfficial is a trusted package page, or a realm an operator aliased
	// that is one. Its head may repeat its leading heading and paragraph.
	pageOfficial
	// pageOperator is a markdown page an operator passed to --aliases. It is
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

	officialImageAsset  = "imgs/og-gnoland.png"
	communityImageAsset = "imgs/og-community.png"
)

// pageLead is what a rendered document says about itself: its leading h1
// and its summary. setHeadMetadata decides whether the head repeats it.
type pageLead struct{ title, description string }

// packageKind classifies u by the package it renders. A user page renders
// that user's home realm. The bare "/r/", "/p/" and "/u/" listings belong to
// no package.
func (h *HTTPHandler) packageKind(u *weburl.GnoURL) pageKind {
	if !(u.IsRealm() || u.IsPure() || u.IsUser()) {
		return pageSite
	}
	switch pkg := u.Path[3:]; { // skip "/r/", "/p/" or "/u/"
	case pkg == "":
		return pageSite
	case h.trusted.contains(pkg):
		return pageOfficial
	default:
		return pageCommunity
	}
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

// links says which links of a document of kind k search engines may follow,
// so a page gno.land does not answer for passes on none of its authority. External links
// stay nofollow on every page but an operator's own: a trusted realm may
// still show what its users wrote.
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

// renderContext is the context a document of kind k renders under.
func (h *HTTPHandler) renderContext(k pageKind) RealmRenderContext {
	return RealmRenderContext{
		ChainId: h.Static.ChainId,
		Remote:  h.Static.RemoteHelp,
		Domain:  h.Static.Domain,
		Links:   k.links(),
	}
}

// defaultDescription is the summary a page gets when it may not, or does not,
// summarise itself.
func (k pageKind) defaultDescription(u *weburl.GnoURL) string {
	switch {
	case k != pageCommunity:
		return siteDescription
	case u.IsPure():
		return communityPackageDescription
	case u.IsUser():
		return communityUserDescription
	default:
		return communityRealmDescription
	}
}

// shareImage is the card image asset. gno.land's plain mark goes only to the
// pages it answers for; a community page gets one marked as community
// content.
func (k pageKind) shareImage() string {
	if k == pageCommunity {
		return communityImageAsset
	}
	return officialImageAsset
}
