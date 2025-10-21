package markdown

import (
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/markdown/extensions"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
	"github.com/yuin/goldmark/parser"
)

// NewGnoParserContext creates a new parser context with GnoURL
func NewGnoParserContext(url *weburl.GnoURL) parser.Context {
	ctx := parser.NewContext()
	ctx.Set(extensions.GnoURLContextKey, *url)
	return ctx
}
