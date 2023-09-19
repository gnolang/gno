package social_feeds

import (
	"encoding/base64"
	"std"
	"strconv"
	"strings"

	"gno.land/r/demo/users"
)

func getFeed(fid FeedID) *Feed {
	fidkey := feedIDKey(fid)
	feed_, exists := gFeeds.Get(fidkey)
	if !exists {
		return nil
	}
	feed := feed_.(*Feed)
	return feed
}

func mustGetFeed(fid FeedID) *Feed {
	feed := getFeed(fid)
	if feed == nil {
		panic("Feed does not exist")
	}
	return feed
}

func incGetFeedID() FeedID {
	gFeedsCtr++
	return FeedID(gFeedsCtr)
}

func usernameOf(addr std.Address) string {
	user := users.GetUserByAddress(addr)
	if user == nil {
		return ""
	} else {
		return user.Name()
	}
}

func feedIDKey(fid FeedID) string {
	return padZero(uint64(fid), 10)
}

func postIDKey(pid PostID) string {
	return padZero(uint64(pid), 10)
}

func padLeft(str string, length int) string {
	if len(str) >= length {
		return str
	} else {
		return strings.Repeat(" ", length-len(str)) + str
	}
}

func padZero(u64 uint64, length int) string {
	str := strconv.Itoa(int(u64))
	if len(str) >= length {
		return str
	} else {
		return strings.Repeat("0", length-len(str)) + str
	}
}

func bytesToString(b []byte) string {
	return base64.RawURLEncoding.EncodeToString(b)
}

func intToString(val int) string {
	return strconv.Itoa(val)
}
