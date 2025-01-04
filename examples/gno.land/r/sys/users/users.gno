package users

import (
	"std"
	"strings"

	"gno.land/p/demo/avl/rotree"
	"gno.land/p/demo/avlhelpers"
)

// ResolveName finds the original NameRecord for a given name or alias.
func ResolveName(name string) *UserData {
	raw, ok := nameStore.Get(name)
	if !ok {
		return nil
	}

	data := raw.(*UserData)
	for data.alias != nil {
		data = data.alias
	}
	return data
}

func ResolveAddress(addr std.Address) *UserData {
	raw, ok := addressStore.Get(addr.String())
	if !ok {
		return nil
	}

	data := raw.(*UserData)
	if data.deleted {
		return nil
	}

	for data.alias != nil {
		data = data.alias
	}
	return data
}

// ResolveAddressOrName
// Note: unlike ResolveName, input must be "@" prefixed for names
// It does not resolve aliases
func ResolveAddressOrName(aon string) *UserData {
	if strings.HasPrefix(aon, "@") {
		return ResolveName(strings.TrimPrefix(aon, "@"))
	}

	return ResolveAddress(std.Address(aon))
}

// GetReadonlyStore exposes the username store in readonly mode
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
