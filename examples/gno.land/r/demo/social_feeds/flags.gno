package social_feeds

import (
	"strconv"
	"strings"

	"gno.land/p/demo/flags_index"
)

var SEPARATOR = "/"

func getFlagID(fid FeedID, pid PostID) flags_index.FlagID {
	return flags_index.FlagID(fid.String() + SEPARATOR + pid.String())
}

func parseFlagID(flagID flags_index.FlagID) (FeedID, PostID) {
	parts := strings.Split(string(flagID), SEPARATOR)
	if len(parts) != 2 {
		panic("invalid flag ID '" + string(flagID) + "'")
	}
	fid, err := strconv.Atoi(parts[0])
	if err != nil || fid == 0 {
		panic("invalid feed ID in flag ID '" + parts[0] + "'")
	}
	pid, err := strconv.Atoi(parts[1])
	if err != nil || pid == 0 {
		panic("invalid post ID in flag ID '" + parts[1] + "'")
	}
	return PostID(fid), PostID(pid)
}
