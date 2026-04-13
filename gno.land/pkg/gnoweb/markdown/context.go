package markdown

import (
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
	"github.com/yuin/goldmark/parser"
)

var (
	gUrlContextKey     = parser.NewContextKey()
	gChainIdContextKey = parser.NewContextKey()
	gRemoteContextKey  = parser.NewContextKey()
	gDomainContextKey  = parser.NewContextKey()
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
