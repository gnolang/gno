// The realm r/system/names is used to manage namespaces on gno.land.
package names

import (
	"std"

	"gno.land/p/demo/avl"
)

// "AddPkg" will check if r/system/names exists. If yes, it will
// inspect the realm's state and use the following variable to
// determine if an address can publish a package or not.
var namespaces avl.Tree // name(string) -> Space

type Space struct {
	Admins  []std.Address
	Editors []std.Address
	InPause bool
}

func Register(namespace string) {
	// TODO: input sanitization:
	// - already exists / reserved.
	// - min/max length, format.
	// - fees (dynamic, based on length).
	if existsNamespace(namespace) {
		panic("namespace already exists")
	}

	caller := std.GetOrigCaller()
	namespaces.Set(namespace, &Space{
		Admins: []std.Address{caller},
	})
}

func AddAdmin(namespace string, newAdmin std.Address) {
	assertIsAdmin(namespace)

	space := getSpace(namespace)
	if !containsAddress(space.Admins, newAdmin) {
		space.Admins = append(space.Admins, newAdmin)
	}
}

func RemoveAdmin(namespace string, newAdmin std.Address) {
	// TODO: check if self.
	assertIsAdmin(namespace)

	// remove admin
	space := getSpace(namespace)
	if len(space.Admins) == 1 {
		panic("namespace at least needs one admin")
	}
	for i, admin := range space.Admins {
		if admin == newAdmin {
			space.Admins = append(space.Admins[:i], space.Admins[i+1:]...)
			break
		}
	}
}

func AddEditor(namespace string, newEditor std.Address) {
	assertIsAdmin(namespace)
	space := getSpace(namespace)

	if !containsAddress(space.Editors, newEditor) {
		space.Editors = append(space.Editors, newEditor)
	}
}

func RemoveEditor(namespace string, newEditor std.Address) {
	assertIsAdmin(namespace)

	// remove editor
	space := getSpace(namespace)
	for i, editor := range space.Editors {
		if editor == newEditor {
			space.Editors = append(space.Editors[:i], space.Editors[i+1:]...)
			break
		}
	}
}

func SetInPause(namespace string, state bool) {
	assertIsAdmin(namespace)
	space := getSpace(namespace)
	space.InPause = state
}

func Render(path string) string {
	// TODO: by namespace.
	// TODO: by address.
	return "not implemented"
}
