package bridge

import (
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
	"gno.land/r/gnoswap/v1/access"
)

var (
	callbackTree       *avl.Tree
	err                error
	adminAddr, govAddr std.Address
)

func init() {
	callbackTree = avl.NewTree()

	adminAddr, err = access.GetAddress(access.ROLE_ADMIN)
	if err != nil {
		panic(err)
	}
	govAddr, err = access.GetAddress(access.ROLE_GOVERNANCE)
	if err != nil {
		panic(err)
	}
}

// RegisterCallback registers a callback function for a given namespace and function name.
// Only admin and governance addresses are allowed to register callbacks.
func RegisterCallback(cur realm, caller std.Address, namespace, fname string, callback any) error {
	if !caller.IsValid() {
		return ufmt.Errorf("invalid caller address: %s", caller)
	}

	// TODO: Currently, Register is only allowed for admin or governance,
	// but for some contracts, the previous caller might not be that contract.
	// Therefore, there is a need to expand permissions.
	//
	// Currently, admin addresses are directly inserted into the caller,
	// but this part is planned to be modified after discussion.
	if caller != adminAddr && caller != govAddr {
		return ufmt.Errorf("unauthorized caller: %s", caller)
	}

	key := createKey(namespace, fname)
	if callbackTree.Has(key) {
		return ufmt.Errorf("callback already registered for key: %s", key)
	}
	callbackTree.Set(key, callback)
	return nil
}

// LookupCallback retrieves a callback function by namespace and function name.
// Returns the callback and a boolean indicating if the callback exists.
func LookupCallback(namespace, fname string) (any, bool) {
	key := createKey(namespace, fname)
	return callbackTree.Get(key)
}

// UpdateCallback updates an existing callback function for a given namespace and function name.
// Only admin and governance addresses are allowed to update callbacks.
func UpdateCallback(cur realm, caller std.Address, namespace, fname string, newCallback any) error {
	if !caller.IsValid() {
		return ufmt.Errorf("invalid caller address: %s", caller)
	}

	// allow only admin and governance
	if caller != adminAddr && caller != govAddr {
		return ufmt.Errorf("unauthorized caller: %s", caller)
	}

	key := createKey(namespace, fname)
	if !callbackTree.Has(key) {
		return ufmt.Errorf("callback not found for key: %s", key)
	}
	callbackTree.Set(key, newCallback)
	return nil
}

// createKey creates a unique key by combining namespace and function name.
func createKey(namespace, fname string) string {
	return ufmt.Sprintf("%s/%s", namespace, fname)
}
