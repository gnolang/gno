package users

import (
	pusers "github.com/gnolang/gno/examples/gno.land/p/demo/users"
	"github.com/gnolang/gno/gnovm/stdlibs/std"
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
func ResolveAddressOrName(input pusers.AddressOrName) *pusers.User {
	name, isName := input.GetName()
	if isName {
		return ResolveUsername(name)
	}

	return ResolveAddress(std.Address(input))
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
