package users

import (
	"github.com/gnolang/gno/examples/gno.land/p/sys/users"
	"std"
)

//type UserLookup struct {
//	Name      string
//	Addr      std.Address
//	UsedAlias bool // lookup was found via an alias
//}

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

// todo refactor
//func ResolveAddress(addr std.Address) *susers.User {
//	if !addr.IsValid() {
//		return nil
//	}
//
//	raw, ok := aStore.Get(addr.String())
//	if !ok {
//		return nil
//	}
//
//	return raw.(*susers.User)
//}

// ResolveAddressOrName
// Note: unlike ResolveName, input must be "@" prefixed for names.
//func ResolveAddressOrName(aon susers.AddressOrName) users.UserLookup {
//	name, isName := aon.Name()
//	if isName {
//		return ResolveName(name)
//	}
//
//	return ResolveAddress(std.Address(aon))
//}
//
//// Users returns a paginated list of users
//func Users(offset, count uint64) []susers.User {
//	if count < 1 {
//		return []susers.User{}
//	}
//
//	res := make([]susers.User, 0, count)
//
//	uStore.IterateByOffset(int(offset), int(count), func(key string, value interface{}) bool {
//		user := value.(susers.User)
//		res = append(res, user)
//
//		return false
//	})
//
//	return res
//}

// XXX: Figure out a flexible but minimalistic API for fetching users from this realm.

// UsersByPrefix gets a slice of usernames starting from the given prefix. Limits the
// number of results to maxResults. (This can be used for a name search tool.)
//func UsersByPrefix(prefix string, maxResults int) []string {
//	// XXX: most likely moved to r/gnoland/users to keep this package minimal.
//	// in that case, should we get a value copy of the uStore available to whitelisted callers?
//	return avlhelpers.ListByteStringKeysByPrefix(*uStore, prefix, maxResults)
//}
