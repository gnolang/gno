package gnotter

import (
	"gno.land/p/avl"
	"std"
	"strconv"
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
