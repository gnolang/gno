package staker

import (
	b64 "encoding/base64"
	"std"
	"strconv"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/json"
	"gno.land/p/demo/ufmt"

	"gno.land/p/gnoswap/consts"
	"gno.land/r/gnoswap/v1/common"
)

// marshal data to json string
func marshal(data *json.Node) string {
	b, err := json.Marshal(data)
	if err != nil {
		panic(err.Error())
	}

	return string(b)
}

// b64Encode encodes data to base64 string
func b64Encode(data string) string {
	return string(b64.StdEncoding.EncodeToString([]byte(data)))
}

// getPrevRealm returns the previous realm's package path
func getPrevRealm() std.Realm {
	return std.PrevRealm()
}

// getPrevAddr returns the address of the previous realm.
func getPrevAddr() std.Address {
	return std.PrevRealm().Addr()
}

// isUserCall returns true if the previous realm is a user
func isUserCall() bool {
	return std.PrevRealm().IsUser()
}

// getPrev returns the previous realm's address and package path
func getPrev() (string, string) {
	prev := getPrevRealm()
	return prev.Addr().String(), prev.PkgPath()
}

// derivePkgAddr derives the Realm address from it's pkgPath parameter
func derivePkgAddr(pkgPath string) std.Address {
	return std.DerivePkgAddr(pkgPath)
}

// formatUint formats a uint64 to a string
func formatUint(v uint64) string {
	return strconv.FormatUint(v, 10)
}

// formatInt formats an int64 to a string
func formatInt(v int64) string {
	return strconv.FormatInt(v, 10)
}

// assertCallerIsAdmin panics if the caller is not an admin
func assertCallerIsAdmin() {
	caller := getPrevAddr()
	if err := common.AdminOnly(caller); err != nil {
		panic(err)
	}
}

// assertCallerIsLaunchpad panics if the caller is not the launchpad
func assertCallerIsLaunchpad() {
	caller := std.PrevRealm().Addr()
	if caller != consts.LAUNCHPAD_ADDR {
		panic(addDetailToError(
			errNoPermission,
			ufmt.Sprintf("only launchpad can call CollectRewardFromLaunchPad(), called from %s", caller.String()),
		))
	}
}

// getUint64FromTree returns the uint64 value from the tree
func getUint64FromTree(tree *avl.Tree, key string) uint64 {
	value, exists := tree.Get(key)
	if !exists {
		return 0
	}

	return value.(uint64)
}

// updateUint64InTree updates the uint64 value in the tree
func updateUint64InTree(tree *avl.Tree, key string, delta uint64, add bool) uint64 {
	current := getUint64FromTree(tree, key)
	var newValue uint64
	if add {
		newValue = current + delta
	} else {
		if current < delta {
			panic(addDetailToError(
				errNotEnoughBalance,
				ufmt.Sprintf("not enough balance: current(%d) < requested(%d)", current, delta),
			))
		}
		newValue = current - delta
	}

	tree.Set(key, newValue)

	return newValue
}

// getOrCreateInnerTree returns the inner tree for the given key
func getOrCreateInnerTree(tree *avl.Tree, key string) *avl.Tree {
	value, exists := tree.Get(key)
	if !exists {
		innerTree := avl.NewTree()
		tree.Set(key, innerTree)
		return innerTree
	}

	return value.(*avl.Tree)
}

// milliToSec converts milliseconds to seconds
func milliToSec(ms int64) int64 {
	return ms / consts.MILLISECONDS_PER_SECOND
}
