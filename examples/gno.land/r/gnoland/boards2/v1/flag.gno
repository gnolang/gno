package boards2

import (
	"strconv"

	"gno.land/p/gnoland/boards"
	"gno.land/p/nt/avl"
)

// DefaultFlaggingThreshold defines the default number of flags that hides flaggable items.
const DefaultFlaggingThreshold = 1

var gFlaggingThresholds avl.Tree // string(board ID) -> int

// flagItem adds a flag to a post.
// Returns whether flag count threshold is reached and post can be hidden.
// Panics if flag count threshold was already reached.
func flagItem(post *boards.Post, user address, reason string, threshold int) bool {
	if post.Flags.Size() >= threshold {
		panic("flag count threshold exceeded: " + strconv.Itoa(threshold))
	}

	if post.Flags.Exists(user) {
		panic("post has been already flagged by " + user.String())
	}

	post.Flags.Add(boards.Flag{
		User:   user,
		Reason: reason,
	})

	return post.Flags.Size() == threshold
}

func getFlaggingThreshold(bid boards.ID) int {
	if v, ok := gFlaggingThresholds.Get(bid.String()); ok {
		return v.(int)
	}
	return DefaultFlaggingThreshold
}
