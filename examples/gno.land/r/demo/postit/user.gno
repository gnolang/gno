package postit

import "std"

type User struct {
	Username  string
	Address   std.Address
	Name      string
	Bio       string
	Posts     []uint64
	Followers []string
	Following []string
	AvatarUrl string // use grc721 instead?
}

func getUserRefByAddress(addr std.Address) *User {
	username, ok := addressToUsername.Get(addr.String())
	if !ok {
		panic("user with addr not found")
	}
	res, ok := users.Get(username.(string))
	if !ok {
		panic("user with username not found")
	}
	return res.(*User)
}

func getUserRefByUsername(username string) *User {
	res, ok := users.Get(username)
	if !ok {
		panic("user not found")
	}
	return res.(*User)
}
