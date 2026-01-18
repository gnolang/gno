package grc20reg

import (
	"chain"
	"chain/runtime"
	"strconv"

	"gno.land/p/demo/tokens/grc20"
	"gno.land/p/nt/avl"
	"gno.land/p/nt/avl/pager"
	"gno.land/p/nt/avl/rotree"
	"gno.land/p/nt/fqname"
	"gno.land/p/nt/mux"
	"gno.land/p/nt/ufmt"
)

var (
	registry      = avl.NewTree() // rlmPath[.slug] -> *Token (slug is optional)
	pagerInstance = pager.NewPager(rotree.Wrap(registry, nil), 10, false)
)

func Register(cur realm, token *grc20.Token, slug string) {
	rlmPath := runtime.PreviousRealm().PkgPath()
	key := fqname.Construct(rlmPath, slug)
	registry.Set(key, token)
	chain.Emit(
		registerEvent,
		"pkgpath", rlmPath,
		"slug", slug,
	)
}

func Get(key string) *grc20.Token {
	token, ok := registry.Get(key)
	if !ok {
		return nil
	}
	return token.(*grc20.Token)
}

func MustGet(key string) *grc20.Token {
	token := Get(key)
	if token == nil {
		panic("unknown token: " + key)
	}
	return token
}

func Render(path string) string {
	router := mux.NewRouter()
	router.HandleFunc("", renderHome)
	router.HandleFunc("*", renderToken)
	return router.Render(path)
}

func renderHome(res *mux.ResponseWriter, req *mux.Request) {
	page := pagerInstance.MustGetPageByPath(req.RawPath)

	if page.TotalItems == 0 {
		res.Write("No registered token.")
		return
	}

	s := ""
	for _, item := range page.Items {
		token := item.Value.(*grc20.Token)
		rlmPath, slug := fqname.Parse(item.Key)
		rlmLink := fqname.RenderLink(rlmPath, slug)
		s += ufmt.Sprintf("- **%s** - %s - [info](/r/demo/defi/grc20reg:%s)\n", token.GetName(), rlmLink, item.Key)
	}

	s += "\n"
	s += "**Page " + strconv.Itoa(page.PageNumber) + " of " + strconv.Itoa(page.TotalPages) + "**\n\n"
	s += page.Picker(req.Path)
	res.Write(s)
}

func renderToken(res *mux.ResponseWriter, req *mux.Request) {
	key := req.Path
	token := MustGet(key)
	rlmPath, slug := fqname.Parse(key)
	rlmLink := fqname.RenderLink(rlmPath, slug)
	s := ufmt.Sprintf("# %s\n", token.GetName())
	s += ufmt.Sprintf("- symbol: **%s**\n", token.GetSymbol())
	s += ufmt.Sprintf("- realm: %s\n", rlmLink)
	s += ufmt.Sprintf("- decimals: %d\n", token.GetDecimals())
	s += ufmt.Sprintf("- total supply: %d\n", token.TotalSupply())
	res.Write(s)
}

const registerEvent = "register"

func GetRegistry() *rotree.ReadOnlyTree {
	return rotree.Wrap(registry, nil)
}
