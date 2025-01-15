package users

import (
	"std"

	"gno.land/p/demo/avl/rotree"
	"gno.land/p/demo/avlhelpers"
)

// ResolveName returns the latest UserData of a specific user by name or Alias
func ResolveName(name string) (data *UserData, deleted bool, isAlias bool) {
	raw, ok := nameStore.Get(name)
	if !ok {
		return nil, false, false
	}

	data = raw.(*UserData)
	return data, data.deleted, name != data.username
}

// ResolveAddress returns the latest UserData of a specific user by address
func ResolveAddress(addr std.Address) (data *UserData, deleted bool) {
	raw, ok := addressStore.Get(addr.String())
	if !ok {
		return nil, false
	}

	data = raw.(*UserData)
	return data, data.deleted
}

//// ResolveActiveAddress resolves an address, but returns nil if the user has been deleted
//func ResolveActiveAddress(addr std.Address) (data *UserData, deleted bool) {
//	raw, ok := addressStore.Get(addr.String())
//	if !ok {
//		return nil, false
//	}
//
//	data = raw.(*UserData)
//	if data.deleted {
//		return nil, true
//	}
//
//	return data, false
//}
//
//// ResolveActiveName resolves a name, but returns nil if the user has been deleted
//func ResolveActiveName(name string) (data *UserData, deleted bool) {
//	raw, ok := nameStore.Get(name)
//	if !ok {
//		return nil, false
//	}
//
//	data = raw.(*UserData)
//	if data.deleted {
//		return nil, true
//	}
//
//	return data, false
//}

// GetReadonlyNameStore exposes the username store in readonly mode
func GetReadonlyNameStore() *rotree.ReadOnlyTree {
	return rotree.Wrap(nameStore, nil)
}

// UsersByPrefix gets a slice of usernames starting from the given prefix. Limits the
// number of results to maxResults. (This can be used for a name search tool.)
func UsersByPrefix(prefix string, maxResults int) []string {
	// XXX: most likely moved to r/gnoland/users to keep this package minimal.
	// in that case, should we get a value copy of the uStore available to whitelisted callers?
	return avlhelpers.ListByteStringKeysByPrefix(nameStore, prefix, maxResults)
}
