package users

import (
	"gno.land/p/demo/avl/rotree"
	"gno.land/p/demo/avlhelpers"
	"std"
)

// ResolveName returns the latest UserData of a specific user by name or Alias
func ResolveName(name string) (data *UserData, deleted bool) {
	raw, ok := nameStore.Get(name)
	if !ok {
		return nil, false
	}

	data = raw.(*UserData)
	for data.Alias != nil {
		data = data.Alias
	}

	return data, data.Deleted
}

// ResolveAddress returns the latest UserData of a specific user by address
func ResolveAddress(addr std.Address) (data *UserData, deleted bool) {
	raw, ok := addressStore.Get(addr.String())
	if !ok {
		return nil, false
	}

	data = raw.(*UserData)
	for data.Alias != nil {
		data = data.Alias
	}

	return data, data.Deleted
}

// ResolveAddressOrName
// Note: unlike ResolveName, input must be "@" prefixed for names
//func ResolveAddressOrName(aon string) *UserData {
//	if strings.HasPrefix(aon, "@") {
//		return ResolveName(strings.TrimPrefix(aon, "@"))
//	}
//
//	return ResolveAddress(std.Address(aon))
//}

// GetReadonlyStore exposes the Username store in readonly mode
func GetReadonlyStore() *rotree.ReadOnlyTree {
	return rotree.Wrap(nameStore, nil)
}

// UsersByPrefix gets a slice of usernames starting from the given prefix. Limits the
// number of results to maxResults. (This can be used for a name search tool.)
func UsersByPrefix(prefix string, maxResults int) []string {
	// XXX: most likely moved to r/gnoland/users to keep this package minimal.
	// in that case, should we get a value copy of the uStore available to whitelisted callers?
	return avlhelpers.ListByteStringKeysByPrefix(nameStore, prefix, maxResults)
}
