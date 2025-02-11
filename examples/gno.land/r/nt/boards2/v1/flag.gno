package boards2

import (
	"std"
	"strconv"

	"gno.land/p/demo/avl"
)

// DefaultFlaggingThreshold defines the default number of flags that hides flaggable items.
const DefaultFlaggingThreshold = 1

var gFlaggingThresholds avl.Tree // string(board ID) -> int

type Flag struct {
	User   std.Address
	Reason string
}

func NewFlag(creator std.Address, reason string) Flag {
	return Flag{
		User:   creator,
		Reason: reason,
	}
}

type Flaggable interface {
	// AddFlag adds a new flag to an item.
	//
	// Returns false if item was already flagged by user.
	AddFlag(flag Flag) bool

	// FlagsCount returns number of times item was flagged.
	FlagsCount() int
}

// flagItem adds a flag to a flaggable item (post, thread, etc).
//
// Returns whether flag count threshold is reached and item can be hidden.
//
// Panics if flag count threshold was already reached.
func flagItem(bid BoardID, item Flaggable, flag Flag) bool {
	threshold := getFlaggingThreshold(bid)
	if item.FlagsCount() >= threshold {
		panic("item flag count threshold exceeded: " + strconv.Itoa(threshold))
	}

	if !item.AddFlag(flag) {
		panic("item has been already flagged by a current user")
	}

	return item.FlagsCount() == threshold
}

func getFlaggingThreshold(bid BoardID) int {
	if v, ok := gFlaggingThresholds.Get(bid.String()); ok {
		return v.(int)
	}
	return DefaultFlaggingThreshold
}
