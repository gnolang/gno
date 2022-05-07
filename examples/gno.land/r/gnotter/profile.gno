package gnotter

import (
	"gno.land/p/avl"
	"std"
	"strconv"
)

const (
	FollowAction   = 0
	UnfollowAction = 1
)

type ProfileID uint64

type Profile struct {
	id        ProfileID
	url       string
	name      string
	creator   std.Address
	threads   *avl.MutTree
	postsCtr  uint64
	createdAt std.Time
	followers *avl.MutTree
	following *avl.MutTree
}

func newProfile(id ProfileID, url string, name string, creator std.Address) *Profile {
	exists := gProfilesByName.Has(name)
	if exists {
		panic("profile already exists")
	}
	return &Profile{
		id:        id,
		url:       url,
		name:      name,
		creator:   creator,
		threads:   avl.NewMutTree(),
		createdAt: std.GetTimestamp(),
		followers: avl.NewMutTree(),
		following: avl.NewMutTree(),
	}
}

func (profile *Profile) RenderProfile() string {
	str := ""
	str += "Name: " + profile.name + "\n"
	str += "URL: " + profile.url + "\n"
	str += "Following: " + strconv.Itoa(profile.following.Size()) + "\n"
	str += "Follower: " + strconv.Itoa(profile.followers.Size()) + "\n"
	str += "Nb threads: " + strconv.Itoa(profile.threads.Size())
	return str
}

func (profile *Profile) FollowAction(action int, profileToTarget ProfileID) {
	profileIdKey := profileIDKey(profileToTarget)
	exists := gProfiles.Has(profileIdKey)
	if !exists {
		panic("profile nonexistent")
	}
	switch action {
	case FollowAction:
		profileToFollow := getProfile(profileToTarget)
		profile.following.Set(profileIdKey, profileToFollow)
		profileToFollow.followers.Set(profileIDKey(profile.id), profile)
		break
	case UnfollowAction:
		profileToUnFollow := getProfile(profileToTarget)
		profile.following.Remove(profileIdKey)
		profileToUnFollow.followers.Remove(profileIDKey(profile.id))
		break
	default:
		panic("unreacheable")
	}
}

func (profile *Profile) Follow(profileToFollowID ProfileID) {
	profile.FollowAction(FollowAction, profileToFollowID)
}

func (profile *Profile) UnFollow(profileToUnfollowID ProfileID) {
	profile.FollowAction(UnfollowAction, profileToUnfollowID)
}
