package markdown

import (
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
	"github.com/yuin/goldmark/parser"
)

var gUrlContextKey = parser.NewContextKey()

// NewGnoParserContext creates a new parser context with GnoURL
func NewGnoParserContext(url *weburl.GnoURL) parser.Context {
	ctx := parser.NewContext()
	ctx.Set(gUrlContextKey, *url)
	return ctx
}

// getUrlFromContext retrieves the GnoURL from the parser context
func getUrlFromContext(ctx parser.Context) (url weburl.GnoURL, ok bool) {
	url, ok = ctx.Get(gUrlContextKey).(weburl.GnoURL)
	return
}
