package social_feeds

import (
	"strconv"
	"strings"
)

func renderFeed(parts []string) string {
	// /r/demo/social_feeds_v3:FEED_NAME
	name := parts[0]
	feedI, exists := gFeedsByName.Get(name)
	if !exists {
		return "feed does not exist: " + name
	}
	return feedI.(*Feed).Render()
}

func renderPost(parts []string) string {
	// /r/demo/boards:FEED_NAME/POST_ID
	name := parts[0]
	feedI, exists := gFeedsByName.Get(name)
	if !exists {
		return "feed does not exist: " + name
	}
	pid, err := strconv.Atoi(parts[1])
	if err != nil {
		return "invalid thread id: " + parts[1]
	}
	feed := feedI.(*Feed)
	post := feed.MustGetPost(PostID(pid))
	return post.Render()
}

func renderFeedsList() string {
	str := "There are " + intToString(gFeeds.Size()) + " available feeds:\n\n"
	gFeeds.Iterate("", "", func(key string, value interface{}) bool {
		feed := value.(*Feed)
		str += " * [" + feed.url + " (FeedID: " + feed.id.String() + ")](" + feed.url + ")\n"
		return false
	})
	return str
}

func Render(path string) string {
	if path == "" {
		return renderFeedsList()
	}

	parts := strings.Split(path, "/")

	if len(parts) == 1 {
		// /r/demo/social_feeds_v3:FEED_NAME
		return renderFeed(parts)
	} else if len(parts) == 2 {
		// /r/demo/social_feeds_v3:FEED_NAME/POST_ID
		return renderPost(parts)
	}
	// } else if len(parts) == 3 {
	// 	// /r/demo/boards:BOARD_NAME/THREAD_ID/REPLY_ID
	// 	name := parts[0]
	// 	feedI, exists := gBoardsByName.Get(name)
	// 	if !exists {
	// 		return "board does not exist: " + name
	// 	}
	// 	pid, err := strconv.Atoi(parts[1])
	// 	if err != nil {
	// 		return "invalid thread id: " + parts[1]
	// 	}
	// 	feed := feedI.(*Feed)
	// 	thread := feed.GetThread(PostID(pid))
	// 	if thread == nil {
	// 		return "thread does not exist with id: " + parts[1]
	// 	}
	// 	rid, err := strconv.Atoi(parts[2])
	// 	if err != nil {
	// 		return "invalid reply id: " + parts[2]
	// 	}
	// 	reply := thread.GetReply(PostID(rid))
	// 	if reply == nil {
	// 		return "reply does not exist with id: " + parts[2]
	// 	}
	// 	return reply.RenderInner()
	// } else {
	// 	return "unrecognized path " + path
	// }
}
