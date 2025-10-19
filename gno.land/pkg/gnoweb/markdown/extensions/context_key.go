package extensions

import (
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
	"github.com/yuin/goldmark/parser"
)

// GnoURLContextKey is the shared context key for storing the GnoURL in Goldmark parser.Context
var GnoURLContextKey = parser.NewContextKey()

// GetUrlFromContext retrieves the GnoURL from the parser context
func GetUrlFromContext(ctx parser.Context) (url weburl.GnoURL, ok bool) {
	if ctx == nil {
		return weburl.GnoURL{}, false
	}

	url, ok = ctx.Get(GnoURLContextKey).(weburl.GnoURL)
	return
}
