package social_follow

import (
	"std"

	"gno.land/p/demo/avl"
)

var addr2User avl.Tree // std.Address -> *User

func getOrCreateUser(addr std.Address) *User {
	userI, ok := addr2User.Get(addr.String())
	if ok {
		return userI.(*User)
	}
	user := &User{
		address: addr,
	}
	addr2User.Set(addr.String(), user)
	return user
}

func Follow(addr std.Address) {
	caller := std.PrevRealm().Addr()
	callerUser := getOrCreateUser(caller)
	user := getOrCreateUser(addr)
	callerUser.Follow(user)
}

func Unfollow(addr std.Address) {
	caller := std.PrevRealm().Addr()
	callerUser := getOrCreateUser(caller)
	user := getOrCreateUser(addr)
	callerUser.Unfollow(user)
}

func Followers(addr std.Address, page, pageSize int) []std.Address {
	userI, ok := addr2User.Get(addr.String())
	if !ok {
		return nil
	}
	user := userI.(*User)
	return user.Followers(page, pageSize)
}

func FollowedCount(addr std.Address) uint {
	userI, ok := addr2User.Get(addr.String())
	if !ok {
		return 0
	}
	user := userI.(*User)
	return uint(user.followeds.Size())
}

func Followed(addr std.Address, page, pageSize int) []std.Address {
	userI, ok := addr2User.Get(addr.String())
	if !ok {
		return nil
	}

	user := userI.(*User)

	return user.Followed(page, pageSize)
}

func FollowersCount(addr std.Address) uint {
	userI, ok := addr2User.Get(addr.String())
	if !ok {
		return 0
	}
	user := userI.(*User)
	return uint(user.followers.Size())
}

func IsFollower(follower std.Address, followed std.Address) bool {
	userI, ok := addr2User.Get(followed.String())
	if !ok {
		return false
	}
	user := userI.(*User)
	_, ok = user.followers.Get(follower.String())
	return ok
}
