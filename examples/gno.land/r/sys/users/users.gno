package users

import (
	"std"

	"gno.land/p/demo/avl/rotree"
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
func ResolveAddress(addr std.Address) *UserData {
	raw, ok := addressStore.Get(addr.String())
	if !ok {
		return nil
	}

	data := raw.(*UserData)
	if data.deleted {
		return nil
	}

	return data
}

// GetReadonlyAddrStore exposes the address store in readonly mode
func GetReadonlyAddrStore() *rotree.ReadOnlyTree {
	return rotree.Wrap(addressStore, makeUserDataSafe)
}

// GetReadOnlyNameStore exposes the name store in readonly mode
func GetReadOnlyNameStore() *rotree.ReadOnlyTree {
	return rotree.Wrap(nameStore, makeUserDataSafe)
}

func makeUserDataSafe(data interface{}) interface{} {
	u := data.(*UserData) // WARN: roTree will return ok = true when getting in this case
	if u.deleted {
		return nil
	}
	return u
}
