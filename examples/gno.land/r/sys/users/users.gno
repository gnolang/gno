package users

import (
	"github.com/gnolang/gno/examples/gno.land/p/demo/avl/rotree"
	"github.com/gnolang/gno/examples/gno.land/p/sys/users"
	"std"
	"strings"

	"gno.land/p/demo/avlhelpers"
)

// ResolveAddressOrName
// Note: unlike ResolveName, input must be "@" prefixed for names
// It does not resolve aliases
func ResolveAddressOrName(aon string) users.UserLookup {
	if strings.HasPrefix(aon, "@") {
		return ResolveName(strings.TrimPrefix(aon, "@"))
	}

	return ResolveAddress(std.Address(aon))
}

// GetReadonlyStore exposes the username store in readonly mode
func GetReadonlyStore() *rotree.ReadOnlyTree {
	return rotree.Wrap(usernameStore, nil)
}

// Users returns a paginated list of user lookups
// Looks users up by username; does not look up aliases
func Users(offset, count uint64) []users.UserLookup {
	if count < 1 {
		return []users.UserLookup{}
	}

	res := make([]users.UserLookup, 0, count)

	usernameStore.IterateByOffset(int(offset), int(count), func(key string, value interface{}) bool {
		res = append(res, users.UserLookup{
			Name:     key,
			Addr:     value.(std.Address),
			ViaAlias: false,
		})

		return false
	})

	return res
}

// UsersByPrefix gets a slice of usernames starting from the given prefix. Limits the
// number of results to maxResults. (This can be used for a name search tool.)
func UsersByPrefix(prefix string, maxResults int) []string {
	// XXX: most likely moved to r/gnoland/users to keep this package minimal.
	// in that case, should we get a value copy of the uStore available to whitelisted callers?
	return avlhelpers.ListByteStringKeysByPrefix(*usernameStore, prefix, maxResults)
}
