package postit

import "std"

func GetPostCount() uint64 {
	return postCount
}

func GetUserCount() uint64 {
	return userCount
}

func GetFeed() {
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
