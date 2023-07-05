package boards

import (
	"strconv"
	"strings"

	"gno.land/p/demo/flags_index"
)

func getFlagID(threadID PostID, postID PostID) flags_index.FlagID {
	return flags_index.FlagID(threadID.String() + "-" + postID.String())
}

func parseFlagID(flagID flags_index.FlagID) (PostID, PostID) {
	parts := strings.Split(string(flagID), "-")
	if len(parts) != 2 {
		panic("invalid flag ID '" + string(flagID) + "'")
	}
	threadID, err := strconv.Atoi(parts[0])
	if err != nil || threadID == 0 {
		panic("invalid thread ID in flag ID '" + parts[0] + "'")
	}
	postID, err := strconv.Atoi(parts[1])
	if err != nil || postID == 0 {
		panic("invalid post ID in flag ID '" + parts[1] + "'")
	}
	return PostID(threadID), PostID(postID)
}
