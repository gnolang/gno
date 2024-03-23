package social_follow

import (
	"std"

	"gno.land/p/demo/avl"
)

type User struct {
	address   std.Address
	followers avl.Tree // std.Address -> *User
	followeds avl.Tree // std.Address -> *User
}

func (u *User) Address() std.Address {
	return u.address
}

func (u *User) Followers(page, pageSize int) []std.Address {
	followers := make([]std.Address, 0, u.followers.Size())
	u.followers.IterateByOffset(page*pageSize, pageSize, func(key string, value interface{}) bool {
		follower := value.(*User)
		followers = append(followers, follower.address)
		return false
	})
	return followers
}

func (u *User) Followed(page, pageSize int) []std.Address {
	followeds := make([]std.Address, 0, u.followeds.Size())
	u.followeds.IterateByOffset(page*pageSize, pageSize, func(key string, value interface{}) bool {
		followed := value.(*User)
		followeds = append(followeds, followed.address)
		return false
	})
	return followeds
}

func (u *User) Follow(user *User) {
	if u.address == user.address {
		panic("can't follow self")
	}
	if _, ok := u.followeds.Get(user.address.String()); ok {
		panic("already follow")
	}
	u.followeds.Set(user.address.String(), user)
	user.followers.Set(u.address.String(), u)
}

func (u *User) Unfollow(user *User) {
	if _, ok := u.followeds.Get(user.address.String()); !ok {
		panic("not follow")
	}
	if _, ok := u.followeds.Remove(user.address.String()); !ok {
		panic("can't remove on followed")
	}

	if _, ok := user.followers.Remove(u.address.String()); !ok {
		panic("can't remove on follower")
	}
}
