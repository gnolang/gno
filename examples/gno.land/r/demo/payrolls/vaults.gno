package payrolls

import (
	"errors"
	"std"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/users"
	rusers "gno.land/r/demo/users"
)

var vaults avl.Tree // std.Addresss -> uint64

func getVaultAmount(id string) uint64 {
	amount, ok := vaults.Get(id)
	if ok {
		return amount.(uint64)
	}
	return 0
}

var errInvalidInput = errors.New("invalid input")

func vaultID(input string, namespace string) (string, error) {
	if len(input) == 0 {
		return "", errInvalidInput
	}

	if id, ok := vaultIDFromPkgPath(input, namespace); ok {
		return id, nil
	}

	user := rusers.GetUserByAddressOrName(users.AddressOrName(input))
	if user != nil {
		return vaultIDUnsafe(user.Address.String(), namespace), nil
	}

	if id, ok := vaultIDFromAddr(std.Address(input), namespace); ok {
		return id, nil
	}

	return "", errInvalidInput
}

func parseVaultID(id string) (string, error) {
	parts := strings.Split(id, ".")
	if len(parts) == 0 {
		return "", errInvalidInput
	}

	if len(parts) == 1 {
		return vaultID(parts[0], "")
	}

	return vaultID(parts[0], parts[1])
}

func vaultIDUnsafe(entity string, namespace string) string {
	if namespace == "" {
		return entity
	}
	return entity + "." + namespace
}

func vaultIDFromAddr(addr std.Address, namespace string) (string, bool) {
	if !addr.IsValid() {
		return "", false
	}
	return vaultIDUnsafe(addr.String(), namespace), true
}

func vaultIDFromPkgPath(entity string, namespace string) (string, bool) {
	if !strings.HasPrefix(entity, std.GetChainDomain()+"/r/") {
		return "", false
	}
	return vaultIDUnsafe(entity, namespace), true
}

func vaultIDFromRealm(realm std.Realm, namespace string) string {
	if realm.IsUser() {
		return vaultIDUnsafe(realm.Addr().String(), namespace)
	} else {
		return vaultIDUnsafe(realm.PkgPath(), namespace)
	}
}
