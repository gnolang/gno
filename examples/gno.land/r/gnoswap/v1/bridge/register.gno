package bridge

import (
	"gno.land/p/demo/avl"
)

var callbackTree *avl.Tree

func init() {
	callbackTree = avl.NewTree()
}

const (
	MINT_AND_DISTRIBUTE_GNS_KEY = "gno.land/r/gnoswap/v1/emission/MintAndDistributeGns"
)

// RegisterCallback registers a callback function with the given key
func RegisterCallback(key string, callback any) {
	if callbackTree.Has(key) {
		panic("callback already registered for key: " + key)
	}
	callbackTree.Set(key, callback)
}

// GetCallback retrieves a callback function by key
func GetCallback(key string) (any, bool) {
	return callbackTree.Get(key)
}
