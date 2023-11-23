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

func ListPostsByOffset(offset, count int) string {
	var pl []Post
	posts.IterateByOffset(offset, count, func(key string, value interface{}) bool {
		post := *value.(*Post)
		pl = append(pl, post)
		return false
	})
	return postsToJSONString(pl)
}

func ListUserPostsByOffset(username string, offset, count int) string {
	var pl []Post
	user := *getUserRefByUsername(username)
	for i, pid := range user.Posts {
		if i < offset {
			continue
		}
		if count <= 0 {
			break
		}
		post, ok := posts.Get(strconv.FormatUint(pid, 10))
		if !ok {
			continue
		}
		pl = append(pl, *post.(*Post))
		count -= 1
	}
	return postsToJSONString(pl)
}
