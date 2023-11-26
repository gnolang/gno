package postit

import (
	"std"
	"strconv"
)

// User

func GetUserCount() uint64 {
	return userCount
}

func GetUserByAddress(addr std.Address) string {
	return userToJSONString(*getUserRefByAddress(addr))
}

func GetUserByUsername(username string) string {
	res, ok := users.Get(username)
	if !ok {
		panic("user not found")
	}
	return userToJSONString(*res.(*User))
}

// Post

func GetPostCount() uint64 {
	return postCount
}

func ListPostsByOffset(offset, limit int) string {
	var pl []Post
	// Ascending
	// for i, pid := range feed {
	// 	if i < offset {
	// 		continue
	// 	}
	// 	if count <= 0 {
	// 		break
	// 	}
	// 	post, ok := posts.Get(strconv.FormatUint(pid, 10))
	// 	if !ok {
	// 		continue
	// 	}
	// 	pl = append(pl, *post.(*Post))
	// 	count -= 1
	// }

	count := 0
	limit = offset + limit
	for i := len(feed) - 1; i >= 0; i-- {
		count++
		if count <= offset {
			continue
		}
		if count > limit {
			break
		}
		post, ok := posts.Get(strconv.FormatUint(feed[i], 10))
		if !ok {
			continue
		}
		pl = append(pl, *post.(*Post))
	}

	return postsToJSONString(pl)
}

func ListUserPostsByOffset(username string, offset, limit int) string {
	var pl []Post
	user := *getUserRefByUsername(username)
	count := 0
	limit = offset + limit
	for i := len(user.Posts) - 1; i >= 0; i-- {
		count++
		if count <= offset {
			continue
		}
		if count > limit {
			break
		}
		post, ok := posts.Get(strconv.FormatUint(user.Posts[i], 10))
		if !ok {
			continue
		}
		pl = append(pl, *post.(*Post))
	}
	return postsToJSONString(pl)
}

// Trending

func GetTrending() string {
	return trendingsToJSONString(trending[:])
}
