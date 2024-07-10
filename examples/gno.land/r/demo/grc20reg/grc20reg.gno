package grc20reg

import (
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/fqname"
	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/ufmt"
)

var registry = avl.NewTree() // rlmPath[.slug] -> *grc20.Bank

func Register(bank *grc20.Bank, slug string) {
	rlmPath := std.PrevRealm().PkgPath()
	key := fqname.Construct(rlmPath, slug)
	registry.Set(key, bank)
	std.Emit(
		registerEvent,
		"pkgpath", rlmPath,
		"slug", slug,
	)
}

func Get(key string) *grc20.Bank {
	bank, ok := registry.Get(key)
	if !ok {
		return nil
	}
	return bank.(*grc20.Bank)
}

func MustGet(key string) *grc20.Bank {
	bank := Get(key)
	if bank == nil {
		panic("unknown bank: " + key)
	}
	return bank
}

func Render(path string) string {
	switch {
	case path == "": // home
		// TODO: add pagination
		s := ""
		count := 0
		registry.Iterate("", "", func(key string, bankI interface{}) bool {
			count++
			bank := bankI.(*grc20.Bank)
			rlmPath, slug := fqname.Parse(key)
			rlmLink := fqname.RenderLink(rlmPath, slug)
			infoLink := "/r/demo/grc20reg:" + key
			s += ufmt.Sprintf("- **%s** - %s - [info](%s)\n", bank.GetName(), rlmLink, infoLink)
			return false
		})
		if count == 0 {
			return "No registered token."
		}
		return s
	default: // specific bank
		key := path
		bank := MustGet(key)
		rlmPath, slug := fqname.Parse(key)
		rlmLink := fqname.RenderLink(rlmPath, slug)
		s := ufmt.Sprintf("# %s\n", bank.GetName())
		s += ufmt.Sprintf("- symbol: **%s**\n", bank.GetSymbol())
		s += ufmt.Sprintf("- realm: %s\n", rlmLink)
		s += ufmt.Sprintf("- decimals: %d\n", bank.GetDecimals())
		s += ufmt.Sprintf("- total supply: %d\n", bank.TotalSupply())
		return s
	}
}

const registerEvent = "register"
