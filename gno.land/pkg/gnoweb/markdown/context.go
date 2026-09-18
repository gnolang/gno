package markdown

import (
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
	"github.com/yuin/goldmark/parser"
)

var (
	gUrlContextKey           = parser.NewContextKey()
	gChainIdContextKey       = parser.NewContextKey()
	gRemoteContextKey        = parser.NewContextKey()
	gDomainContextKey        = parser.NewContextKey()
	gForeignOriginContextKey = parser.NewContextKey()
)

type GnoContext struct {
	GnoURL  *weburl.GnoURL
	ChainId string
	Remote  string
	Domain  string
}

// NewGnoParserContext creates a new parser context with GnoURL
func NewGnoParserContext(mdctx GnoContext) parser.Context {
	ctx := parser.NewContext()
	ctx.Set(gUrlContextKey, mdctx.GnoURL)
	ctx.Set(gChainIdContextKey, mdctx.ChainId)
	ctx.Set(gRemoteContextKey, mdctx.Remote)
	ctx.Set(gDomainContextKey, mdctx.Domain)
	return ctx
}

// getGnoContext reconstructs the GnoContext set by NewGnoParserContext.
// Used to carry the render context across the <gno-foreign> inner-
// instance boundary so links/extensions inside a sandbox get the same
// (URL-aware, dangerous-URL-guarded) treatment as top-level content.
func getGnoContext(ctx parser.Context) GnoContext {
	url, _ := getUrlFromContext(ctx)
	chainId, _ := getChainIdFromContext(ctx)
	remote, _ := getRemoteFromContext(ctx)
	domain, _ := getDomainFromContext(ctx)
	return GnoContext{GnoURL: url, ChainId: chainId, Remote: remote, Domain: domain}
}

// getUrlFromContext retrieves the GnoURL from the parser context
func getUrlFromContext(ctx parser.Context) (url *weburl.GnoURL, ok bool) {
	if url, ok = ctx.Get(gUrlContextKey).(*weburl.GnoURL); url == nil {
		return nil, false
	}

	return
}

// getChainIdFromContext retrieves the ChainId from the parser context
func getChainIdFromContext(ctx parser.Context) (chainId string, ok bool) {
	if chainId, ok = ctx.Get(gChainIdContextKey).(string); !ok {
		return "", false
	}
	return
}

// getRemoteFromContext retrieves the Remote from the parser context
func getRemoteFromContext(ctx parser.Context) (remote string, ok bool) {
	if remote, ok = ctx.Get(gRemoteContextKey).(string); !ok {
		return "", false
	}
	return
}

// getDomainFromContext retrieves the Domain from the parser context
func getDomainFromContext(ctx parser.Context) (domain string, ok bool) {
	if domain, ok = ctx.Get(gDomainContextKey).(string); !ok {
		return "", false
	}
	return
}

// markForeignOrigin flags ctx as the parser context of a <gno-foreign>
// sandbox INNER instance. It is set only by the foreign renderer (see
// ext_foreign.go renderForeign), never by NewGnoParserContext — the
// public GnoContext carries no notion of trust, and a top-level page
// must never be treated as foreign.
//
// Links parsed under a flagged context render as untrusted user-
// generated content: rel="noopener nofollow ugc" and no first-party
// tx/internal trust icons (see linkTransformer.Transform and
// getLinkIcons), so sandboxed foreign markdown cannot wear the host
// realm's link chrome.
func markForeignOrigin(ctx parser.Context) {
	ctx.Set(gForeignOriginContextKey, true)
}

// isForeignOrigin reports whether ctx was flagged by markForeignOrigin.
func isForeignOrigin(ctx parser.Context) bool {
	v, _ := ctx.Get(gForeignOriginContextKey).(bool)
	return v
}
