package users

import (
	"chain"

	"gno.land/p/demo/avl/rotree"
)

// ResolveName returns the latest UserData of a specific user by name or alias
func ResolveName(name string) (data *UserData, isCurrent bool) {
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
func ResolveAddress(addr chain.Address) *UserData {
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

func makeUserDataSafe(data any) any {
	cpy := new(UserData)
	*cpy = *(data.(*UserData))
	if cpy.deleted {
		return nil
	}

	// Note: when requesting data from this AVL tree, (exists bool) will be true
	// Even if the data is "deleted". This is currently unavoidable
	return cpy
}
