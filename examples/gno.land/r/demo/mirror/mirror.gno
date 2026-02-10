package mirror

import (
	"gno.land/p/nt/avl"
)

var store avl.Tree

func Register(_ realm, pkgpath string, rndr func(string) string) {
	if store.Has(pkgpath) {
		return
	}

	if rndr == nil {
		return
	}

	store.Set(pkgpath, rndr)
}

func Render(path string) string {
	if raw, ok := store.Get(path); ok {
		return raw.(func(string) string)("")
	}

	if store.Size() == 0 {
		return "None are fair."
	}

	return "Mirror, mirror on the wall, which realm's the fairest of them all?"
}

// Credits to @jeronimoalbi
