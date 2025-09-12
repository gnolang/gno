package markdown

import (
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/yuin/goldmark/parser"
)

var (
	gUrlContextKey        = parser.NewContextKey()
	gRealmFuncsContextKey = parser.NewContextKey()
)

type RealmFuncSigGetter func(fn string) (*vm.FunctionSignature, error)

type GnoContext struct {
	GnoURL             *weburl.GnoURL
	RealmFuncSigGetter RealmFuncSigGetter
}

// NewGnoParserContext creates a new parser context with GnoURL
func NewGnoParserContext(mdctx GnoContext) parser.Context {
	ctx := parser.NewContext()
	ctx.Set(gUrlContextKey, mdctx.GnoURL)
	ctx.Set(gRealmFuncsContextKey, mdctx.RealmFuncSigGetter)
	return ctx
}

// getUrlFromContext retrieves the GnoURL from the parser context
func getUrlFromContext(ctx parser.Context) (url *weburl.GnoURL, ok bool) {
	if url, ok = ctx.Get(gUrlContextKey).(*weburl.GnoURL); url == nil {
		return nil, false
	}

	return
}

// getRealmFuncsGetter retrieves the Sigs of the given function
func getRealmFuncsGetterFromContext(ctx parser.Context) (gfuncs RealmFuncSigGetter, ok bool) {
	if gfuncs, ok = ctx.Get(gRealmFuncsContextKey).(RealmFuncSigGetter); gfuncs == nil {
		return nil, false
	}

	return
}
