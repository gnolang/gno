package gnoweb

import "context"

// The head's fixed wording, for the external tests to assert against.
const (
	// SiteDescription is the summary of an official page without one, on gno.land.
	SiteDescription  = "Explore realms and packages on gno.land, the network for Gno smart contracts."
	ErrorDescription = errorDescription
)

// MaxUsernameLen exposes the gate's cap on a /u/ name to the external tests.
const MaxUsernameLen = maxUsernameLen

// SitemapLocs exposes what the sitemap lists to the external tests, so they
// can open each page it names. No route of the router's own shadows an alias.
func (h *HTTPHandler) SitemapLocs(ctx context.Context) []string {
	locs, _ := h.sitemapLocs(ctx, func(string) bool { return true })
	return locs
}

// ResolveAlias returns t with a static page's file as it is on disk now.
func ResolveAlias(t AliasTarget) AliasTarget { return t.resolve() }
