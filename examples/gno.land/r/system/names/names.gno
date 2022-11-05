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

func (s *Space) isAdmin(addr std.Address) bool {
	if containsAddress(s.Admins, addr) {
		return true
	}
	return false
}

func (s *Space) addAdmin(newAdmin std.Address) {
	if !containsAddress(s.Admins, newAdmin) {
		s.Admins = append(s.Admins, newAdmin)
	}
}

func (s *Space) removeAdmin(admin std.Address) error {
	if len(s.Admins) == 1 {
		return errors.New("namespace at least needs one admin")
	}
	admins := removeAddress(s.Admins, admin)
	s.Admins = admins
	return nil
}

func (s *Space) addEditor(newEditor std.Address) {
	if !containsAddress(s.Editors, newEditor) {
		s.Editors = append(s.Editors, newEditor)
	}
}

func (s *Space) removeEditor(editor std.Address) error {
	editors := removeAddress(s.Editors, editor)
	s.Editors = editors
	return nil
}

func (s *Space) hasPerm(caller std.Address) bool {
	if s.InPause {
		return false
	}

	if containsAddress(s.Admins, caller) {
		return true
	}

	if containsAddress(s.Editors, caller) {
		return true
	}

	return false
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
	space := getSpace(namespace)
	assertIsAdmin(space)
	space.addAdmin(newAdmin)
}

func RemoveAdmin(namespace string, admin std.Address) {
	space := getSpace(namespace)
	assertIsAdmin(space)
	err := space.removeAdmin(admin)
	checkErr(err)
}

func AddEditor(namespace string, newEditor std.Address) {
	space := getSpace(namespace)
	assertIsAdmin(space)
	space.addEditor(newEditor)
}

func RemoveEditor(namespace string, editor std.Address) {
	space := getSpace(namespace)
	assertIsAdmin(space)
	err := space.removeEditor(editor)
	checkErr(err)
}

func SetInPause(namespace string, state bool) {
	space := getSpace(namespace)
	assertIsAdmin(space)
	space.InPause = state
}

// HasPerm returns true if the caller has permission of the namespace.
// If the namespace does not exist, it will return panic.
// If the namespace exists but the caller is not an admin or editor,
// it will return false.
// The vm keeper will use this function to check to add package
func HasPerm(namespace string) bool {
	caller := std.GetOrigCaller()
	space := getSpace(namespace)
	return space.hasPerm(caller)
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
