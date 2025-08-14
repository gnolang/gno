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
