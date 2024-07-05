package grc20_registry

import (
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/fqname"
	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/ufmt"
)

var registry = avl.NewTree() // pkg path -> IGRC20

func Register(token grc20.IGRC20, slug string) {
	rlmPath := std.PrevRealm().PkgPath()
	key := fqname.Construct(rlmPath, slug)
	registry.Set(key, token)
}

func Get(key string) grc20.IGRC20 {
	token, ok := registry.Get(key)
	if !ok {
		return nil
	}
	return token.(grc20.IGRC20)
}

func MustGet(key string) grc20.IGRC20 {
	token := Get(key)
	if token == nil {
		panic("unknown token: " + key)
	}
	return token
}

func Render(path string) string {
	s := ""

	// TODO: add pagination
	count := 0
	registry.Iterate("", "", func(key string, tokenI interface{}) bool {
		count++
		token := tokenI.(grc20.IGRC20)
		rlmPath, slug := fqname.Parse(key)
		rlmLink := fqname.RenderLink(rlmPath, slug)
		s += ufmt.Sprintf("- **%s** - %s\n", token.GetName(), rlmLink)
		return false
	})
	if count == 0 {
		return "No registered token."
	}
	return s
}
