package users

import (
	"std"

	"gno.land/p/demo/avlhelpers"
	pusers "gno.land/p/demo/users"
)

func ResolveUsername(username string) *pusers.User {
	if username == "" {
		return nil
	}

	raw, ok := uStore.Get(username)
	if !ok {
		return nil
	}

	return raw.(*pusers.User)
}

func ResolveAddress(addr std.Address) *pusers.User {
	if !addr.IsValid() {
		return nil
	}

	raw, ok := aStore.Get(addr.String())
	if !ok {
		return nil
	}

	return raw.(*pusers.User)
}

// ResolveAddressOrName
// Note: unlike ResolveUsername, input must be "@" prefixed for names.
func ResolveAddressOrName(aon pusers.AddressOrName) *pusers.User {
	name, isName := aon.Name()
	if isName {
		return ResolveUsername(name)
	}

	return ResolveAddress(std.Address(aon))
}

// Users returns a paginated list of users
func Users(offset, count uint64) []pusers.User {
	if count < 1 {
		return []pusers.User{}
	}

	res := make([]pusers.User, 0, count)

	uStore.IterateByOffset(int(offset), int(count), func(key string, value interface{}) bool {
		user := value.(pusers.User)
		res = append(res, user)

		return false
	})

	return res
}

// UsersByPrefix gets a slice of usernames starting from the given prefix. Limits the
// number of results to maxResults. (This can be used for a name search tool.)
func UsersByPrefix(prefix string, maxResults int) []string {
	// most likely moved to r/gnoland/users to keep this package minimal.
	// in that case, should we get a value copy of the uStore available to whitelisted callers?
	return avlhelpers.ListByteStringKeysByPrefix(*uStore, prefix, maxResults)
}

// UpdateBio allows the caller to change their bio
func UpdateBio(bio string) error {
	caller := std.GetOrigCaller()
	u := ResolveAddress(caller)
	if u == nil {
		return ErrNotRegistered
	}

	u.Bio = bio
	uStore.Set(u.Name, u)

	return nil
}
