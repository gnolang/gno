package names

import (
	"std"

	"gno.land/p/demo/ufmt"
)

func Register(namespace string) {
	// TODO: input sanitization:
	// - already exists / reserved.
	// - min/max length, format.
	// - fees (dynamic, based on length).
	if existsNamespace(namespace) {
		panic("namespace already exists")
	}

	err := validateNamespace(namespace)
	checkErr(err)

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
	// if enabled is false, it always returns true for dev and testing.
	if !enabled {
		return true
	}
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
	namespaces.Iterate("", "", func(namespace string, value interface{}) bool {
		space := value.(*Space)
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
