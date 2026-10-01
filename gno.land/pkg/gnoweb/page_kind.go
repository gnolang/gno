package gnoweb

import "github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"

// pageKind says who answers for a page, which decides what its <head> may
// repeat. gno.land is permissionless, so a title, a summary or a share card
// lifted from any realm would let anyone speak under gno.land's name (#3910).
// The zero value is the most restrictive, so an unclassified page fails closed.
type pageKind int

const (
	// pageCommunity is a package or user page outside the trusted paths. Its
	// head repeats nothing it renders.
	pageCommunity pageKind = iota
	// pageOfficial is a page an operator aliased or a trusted package page.
	// Its head may repeat its first heading and its first paragraph.
	pageOfficial
	// pageSite is a gnoweb view that belongs to no package, such as the bare
	// "/r/" listing. Its head is generic.
	pageSite
)

const (
	siteDescription             = "Explore realms and packages on gno.land, the network for Gno smart contracts."
	communityRealmDescription   = "A community realm on gno.land, published by its author and not reviewed by the gno.land team."
	communityPackageDescription = "A community package on gno.land, published by its author and not reviewed by the gno.land team."
	communityUserDescription    = "A gno.land user profile. Its content is not reviewed by the gno.land team."

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
	return k == pageOfficial && u.Args == "" && len(u.Query) == 0
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
// pages it answers for; a community page gets one that says it is not.
func (k pageKind) shareImage() string {
	if k == pageCommunity {
		return communityImageAsset
	}
	return officialImageAsset
}
