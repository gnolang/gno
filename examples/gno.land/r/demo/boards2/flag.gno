package boards2

import (
	"std"
	"strconv"
)

var flagThreshold = 3

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
	// AddFlag adds a new flag to an entry.
	//
	// Returns false if item was already flagged by user.
	AddFlag(flag Flag) bool

	// FlagsCount returns number of times entry was flagged.
	FlagsCount() int
}

// flagItem adds a flag to a flaggable item (post, thread, etc).
//
// Returns whether flag count threshold is reached and item can be hidden.
//
// Panics if flag count threshold was already reached.
func flagItem(item Flaggable, flag Flag) bool {
	if item.FlagsCount() >= flagThreshold {
		panic("item flag count threshold exceeded: " + strconv.Itoa(flagThreshold))
	}

	if !item.AddFlag(flag) {
		panic("item has been already flagged by a user")
	}

	return item.FlagsCount() == flagThreshold
}
