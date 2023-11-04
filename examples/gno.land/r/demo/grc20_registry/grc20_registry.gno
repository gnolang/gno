package grc20_registry

import (
	"std"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/grc/grc20"
)

var registry = avl.NewTree() // pkg path -> IGRC20

func Register(token grc20.IGRC20) {
	caller := std.PrevRealm().PkgPath()
	registry.Set(caller, token)
}

func Get(pkgPath string) (grc20.IGRC20, bool) {
	coinI, ok := registry.Get(pkgPath)
	if !ok {
		return nil, false
	}
	coin, ok := coinI.(grc20.IGRC20)
	if !ok {
		panic("internal error: registered object is not a GRC20 token")
	}
	return coin, true
}

func Render(path string) string {
	s := "# GRC20 Registry\n\n" +
		"## Registered Tokens\n\n"
	registry.Iterate("", "", func(pkgPath string, tokenI interface{}) bool {
		token, ok := tokenI.(grc20.IGRC20)
		pkgWebPath := strings.TrimPrefix(pkgPath, "gno.land")
		if ok {
			s += "- [" + token.GetName() + " (" + pkgPath + ")](" + pkgWebPath + ")\n"
		} else {
			s += "- [internal error: registered object is not a GRC20 token (" + pkgPath + ")](" + pkgWebPath + ")\n"
		}
		return false
	})
	return s
}
