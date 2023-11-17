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
	username, ok := addressToUsername.Get(addr.String())
	if !ok {
		panic("user not found")
	}
	res, ok := users.Get(username.(string))
	if !ok {
		panic("user not found")
	}
	return *res.(*User)
}

func GetUserByUsername(username string) User {
	res, ok := users.Get(username)
	if !ok {
		panic("user not found")
	}
	return *res.(*User)
}
