package registry

import (
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
)

var tree = avl.NewTree()

func Register() {
	crossing()
	if std.PreviousRealm().IsUser() {
		return
	}

	if tree.Has(std.PreviousRealm().PkgPath()) {
		return
	}

	tree.Set(std.PreviousRealm().PkgPath(), std.PreviousRealm().Address())
}

func Render(_ string) string {
	out := "# Registry\n\n"

	tree.Iterate("", "", func(key string, value any) bool {
		out += ufmt.Sprintf("%s: %s", key, value.(string))
		return false
	})

	return out
}
