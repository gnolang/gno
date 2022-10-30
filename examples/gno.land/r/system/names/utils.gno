package names

import (
	"std"
)

func getSpace(namespace string) Space {
	s, ok := namespaces.Get(namespace)
	if !ok {
		panic("namespace does not exist")
	}
	return s.(Space)
}

func existsNamespace(name string) bool {
	_, ok := namespaces.Get(name)
	return ok
}

func assertIsAdmin(namespace string) {
	caller := std.GetOrigCaller()
	space := getSpace(namespace)

	if !containsAddress(space.Admins, caller) {
		panic("Only admins can call this function")
	}
}

func containsAddress(addrs []std.Address, addr std.Address) bool {
	for _, a := range addrs {
		if a == addr {
			return true
		}
	}
	return false
}
