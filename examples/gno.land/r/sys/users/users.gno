package users

import (
	"std"

	"gno.land/p/demo/avl/rotree"
	"gno.land/p/demo/avlhelpers"
)

// ResolveName returns the latest UserData of a specific user by name or alias
func ResolveName(name string) (data *UserData, isLatest bool) {
	raw, ok := nameStore.Get(name)
	if !ok {
		return nil, false
	}

	data = raw.(*UserData)
	if data.deleted {
		return nil, false
	}

	return data, name == data.username
}

// ResolveAddress returns the latest UserData of a specific user by address
func ResolveAddress(addr std.Address) (data *UserData) {
	raw, ok := addressStore.Get(addr.String())
	if !ok {
		return nil
	}

	data = raw.(*UserData)
	if data.deleted {
		return nil
	}

	return data
}

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
