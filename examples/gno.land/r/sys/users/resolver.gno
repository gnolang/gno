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
