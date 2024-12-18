package users

import (
	"github.com/gnolang/gno/examples/gno.land/p/demo/avl/rotree"
	"github.com/gnolang/gno/examples/gno.land/p/sys/users"
	"std"
	"strings"

	"gno.land/p/demo/avlhelpers"
)

func ResolveName(username string) users.UserLookup {
	if username == "" {
		return users.UserLookup{}
	}

	// Try fetching from username store
	if rawAddr, ok := usernameStore.Get(username); ok {
		return users.NewUserLookup(username, rawAddr.(std.Address), false)
	}

	// Try fetching from alias store if not found in username store
	if rawAddr, ok := aliasStore.Get(username); ok {
		return users.NewUserLookup(username, rawAddr.(std.Address), true)
	}

	// Not found in either store
	return users.UserLookup{}
}

// ResolveAddress resolves a specific address' username
// It does not resolve aliases
func ResolveAddress(addr std.Address) users.UserLookup {
	if !addr.IsValid() {
		return users.UserLookup{}
	}

	raw, ok := addressStore.Get(addr.String())
	if !ok {
		return users.UserLookup{}
	}

	return users.NewUserLookup(raw.(string), addr, false)
}

// ResolveAddressOrName
// Note: unlike ResolveName, input must be "@" prefixed for names
// It does not resolve aliases
func ResolveAddressOrName(aon string) users.UserLookup {
	if strings.HasPrefix(aon, "@") {
		return ResolveName(strings.TrimPrefix(aon, "@"))
	}

	return ResolveAddress(std.Address(aon))
}

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
		res = append(res, users.NewUserLookup(key, value.(std.Address), false))
		return false
	})

	return res
}

// GetAllNames gets the name and all aliases for a specific address
func GetAllNames(addr std.Address) []string {
	var res []string

	username, ok := addressStore.Get(addr.String())
	if !ok {
		return res
	}

	res = append(res, username.(string))

	aliasStore.Iterate("", "", func(key string, value interface{}) bool {
		adr := value.(std.Address)
		if adr == addr {
			res = append(res, key)
		}

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
