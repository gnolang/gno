// The realm r/system/names is used to manage namespaces on gno.land.
package names

import (
	"errors"
	"regexp"
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
)

// "AddPkg" will check if r/system/names exists. If yes, it will
// inspect the realm's state and use the following variable to
// determine if an address can publish a package or not.
var namespaces avl.Tree // name(string) -> Space

// TODO: more accurate.
var reNamespace = regexp.MustCompile(`^[a-z][a-z0-9_]{2,30}$`)

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

	if err := validateNamespace(namespace); err != nil {
		panic(err)
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
	var admins []std.Address
	for i, admin := range space.Admins {
		if admin == newAdmin {
			admins = append(space.Admins[:i], space.Admins[i+1:]...)
			break
		}
	}
	space.Admins = admins
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
	var editors []std.Address
	for i, editor := range space.Editors {
		if editor == newEditor {
			editors = append(space.Editors[:i], space.Editors[i+1:]...)
			break
		}
	}
	space.Editors = editors
}

func SetInPause(namespace string, state bool) {
	assertIsAdmin(namespace)
	space := getSpace(namespace)
	space.InPause = state
}

// HasPerm returns true if the caller has permission of the namespace.
func HasPerm(namespace string) bool {
	caller := std.GetOrigCaller()
	space := getSpace(namespace)

	if space.InPause {
		return false
	}

	if containsAddress(space.Admins, caller) {
		return true
	}

	if containsAddress(space.Editors, caller) {
		return true
	}

	return false
}

func Render(path string) string {
	// TODO: by address.

	if path == "" {
		return renderIndex()
	} else if path[:2] == "n/" {
		return renderNamespace(path[2:])
	}
	return ""
}

func renderNamespace(namespace string) string {
	space := getSpace(namespace)
	output := ufmt.Sprintf(`
# %s

## Admins

%s

## Editors

%s

## InPause

%s

`, namespace, renderAddresses(space.Admins), renderAddresses(space.Editors), formatBool(space.InPause))
	return output

}

func renderIndex() string {
	output := "## Namespaces \n"
	namespaces.Iterate("", "", func(n *avl.Tree) bool {
		namespace := n.Key()
		space := n.Value().(*Space)
		output += ufmt.Sprintf("* [%s](/r/system/names:n/%s) -	admins: %d editors: %d inPause: %s \n",
			namespace, namespace, len(space.Admins), len(space.Editors), formatBool(space.InPause))
		return false
	})

	return output
}

func renderAddresses(addresses []std.Address) string {
	output := ""
	for _, address := range addresses {
		output += ufmt.Sprintf("* %s \n", string(address))
	}
	return output
}

func validateNamespace(namespace string) error {
	if !reNamespace.MatchString(namespace) {
		return errors.New("invalid namespace")
	}
	return nil
}
