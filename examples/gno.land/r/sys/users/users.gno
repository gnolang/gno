package users

import (
	"std"

	"gno.land/p/demo/avlhelpers"
	susers "gno.land/p/sys/users"
)

func ResolveName(username string) *susers.User {
	if username == "" {
		return nil
	}

	raw, ok := uStore.Get(username)
	if !ok {
		return nil
	}

	return raw.(*susers.User)
}

func ResolveAddress(addr std.Address) *susers.User {
	if !addr.IsValid() {
		return nil
	}

	raw, ok := aStore.Get(addr.String())
	if !ok {
		return nil
	}

	return raw.(*susers.User)
}

// ResolveAddressOrName
// Note: unlike ResolveName, input must be "@" prefixed for names.
func ResolveAddressOrName(aon susers.AddressOrName) *susers.User {
	name, isName := aon.Name()
	if isName {
		return ResolveName(name)
	}

	return ResolveAddress(std.Address(aon))
}

// Users returns a paginated list of users
func Users(offset, count uint64) []susers.User {
	if count < 1 {
		return []susers.User{}
	}

	res := make([]susers.User, 0, count)

	uStore.IterateByOffset(int(offset), int(count), func(key string, value interface{}) bool {
		user := value.(susers.User)
		res = append(res, user)

		return false
	})

	return res
}

// XXX: Figure out a flexible but minimalistic API for fetching users from this realm.

// UsersByPrefix gets a slice of usernames starting from the given prefix. Limits the
// number of results to maxResults. (This can be used for a name search tool.)
func UsersByPrefix(prefix string, maxResults int) []string {
	// XXX: most likely moved to r/gnoland/users to keep this package minimal.
	// in that case, should we get a value copy of the uStore available to whitelisted callers?
	return avlhelpers.ListByteStringKeysByPrefix(*uStore, prefix, maxResults)
}
