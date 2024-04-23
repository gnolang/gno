// The realm r/sys/names is used to manage namespaces on gno.land.
package names

import (
	"std"

	"gno.land/p/demo/avl"
)

// "AddPkg" will check if r/sys/names exists. If yes, it will
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
	panic("not implemented")
}

func AddAdmin(namespace string, newAdmin std.Address) {
	// TODO: assertIsAdmin()
	panic("not implemented")
}

func RemoveAdmin(namespace string, newAdmin std.Address) {
	// TODO: assertIsAdmin()
	// TODO: check if self.
	panic("not implemented")
}

func AddEditor(namespace string, newAdmin std.Address) {
	// TODO: assertIsAdmin()
	panic("not implemented")
}

func RemoveEditor(namespace string, newAdmin std.Address) {
	// TODO: assertIsAdmin()
	// TODO: check if self.
	panic("not implemented")
}

func SetInPause(namespace string, state bool) {
	// TODO: assertIsAdmin()
	panic("not implemented")
}

func Render(path string) string {
	// TODO: by namespace.
	// TODO: by address.
	return "not implemented"
}
