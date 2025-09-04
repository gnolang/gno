package boards2

import (
	"chain"
	"strconv"

	"gno.land/p/demo/avl"
)

// DefaultFlaggingThreshold defines the default number of flags that hides flaggable items.
const DefaultFlaggingThreshold = 1

var gFlaggingThresholds avl.Tree // string(board ID) -> int

type Flaggable interface {
	// Flag adds a new flag to an item.
	// It returns false when the user flagging already flagged the item.
	Flag(user chain.Address, reason string) bool

	// FlagsCount returns number of times item was flagged.
	FlagsCount() int
}

// flagItem adds a flag to a flaggable item.
// Returns whether flag count threshold is reached and item can be hidden.
// Panics if flag count threshold was already reached.
func flagItem(item Flaggable, user chain.Address, reason string, threshold int) bool {
	if item.FlagsCount() >= threshold {
		panic("item flag count threshold exceeded: " + strconv.Itoa(threshold))
	}

	if !item.Flag(user, reason) {
		panic("item has been already flagged by " + user.String())
	}

	return item.FlagsCount() == threshold
}

func getFlaggingThreshold(bid BoardID) int {
	if v, ok := gFlaggingThresholds.Get(bid.String()); ok {
		return v.(int)
	}
	return DefaultFlaggingThreshold
}
