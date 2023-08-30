package names

import (
	"std"
)

func getSpace(namespace string) *Space {
	s, ok := namespaces.Get(namespace)
	if !ok {
		panic("namespace does not exist")
	}
	return s.(*Space)
}

func existsNamespace(name string) bool {
	_, ok := namespaces.Get(name)
	return ok
}

func assertIsAdmin(space *Space) {
	caller := std.GetOrigCaller()
	if !space.isAdmin(caller) {
		panic("Only admins can call this function")
	}
}

func removeAddress(addrs []std.Address, addr std.Address) []std.Address {
	var newAddrs []std.Address
	for _, a := range addrs {
		if a != addr {
			newAddrs = append(newAddrs, a)
		}
	}
	return newAddrs
}

func containsAddress(addrs []std.Address, addr std.Address) bool {
	for _, a := range addrs {
		if a == addr {
			return true
		}
	}
	return false
}

func isCallerAddress(addr std.Address) bool {
	if addr == std.GetOrigCaller() {
		return true
	}
	return false
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func formatBool(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
