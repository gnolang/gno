package payrolls

import (
	"chain"
	"chain/runtime"
	"errors"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/users"
	rusers "gno.land/r/demo/users"
)

var vaults avl.Tree // std.Addresss -> std.Coins

var errInvalidInput = errors.New("invalid input")

func getVaultCoins(id string) chain.Coins {
	amount, ok := vaults.Get(id)
	if ok {
		return amount.(chain.Coins)
	}
	return nil
}

func assertValidNamespace(namespace string) {
	dotIdx := strings.Index(namespace, ".")
	if dotIdx != -1 {
		panic(errors.New("a namespace must not contain dots"))
	}
}

func vaultID(base string, namespace string) (string, error) {
	if len(base) == 0 {
		return "", errInvalidInput
	}

	if id, ok := vaultIDFromPkgPath(base, namespace); ok {
		return id, nil
	}

	user := rusers.GetUserByAddressOrName(users.AddressOrName(base))
	if user != nil {
		return vaultIDUnsafe(user.Address.String(), namespace), nil
	}

	if id, ok := vaultIDFromAddr(address(base), namespace); ok {
		return id, nil
	}

	return "", errInvalidInput
}

func splitVaultInput(input string) (string, string) {
	lastDot := strings.LastIndex(input, ".")
	if lastDot == -1 {
		return input, ""
	}

	base := input[:lastDot]
	namespace := input[lastDot+1:]
	return base, namespace
}

func vaultIDUnsafe(base string, namespace string) string {
	if namespace == "" {
		return base
	}
	return base + "." + namespace
}

func vaultIDFromAddr(addr address, namespace string) (string, bool) {
	if !addr.IsValid() {
		return "", false
	}
	return vaultIDUnsafe(addr.String(), namespace), true
}

func vaultIDFromPkgPath(base string, namespace string) (string, bool) {
	if !strings.HasPrefix(base, runtime.ChainDomain()+"/r/") {
		return "", false
	}
	return vaultIDUnsafe(base, namespace), true
}

func vaultIDFromRealm(base runtime.Realm, namespace string) string {
	if base.IsUser() {
		return vaultIDUnsafe(base.Addr().String(), namespace)
	} else {
		return vaultIDUnsafe(base.PkgPath(), namespace)
	}
}
