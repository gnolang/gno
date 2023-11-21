package postit

import "std"

// User

func GetUserCount() uint64 {
	return userCount
}

func GetUserByAddress(addr std.Address) User {
	return *getUserRefByAddress(addr)
}

func GetUserByUsername(username string) User {
	res, ok := users.Get(username)
	if !ok {
		panic("user not found")
	}
	return *res.(*User)
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
